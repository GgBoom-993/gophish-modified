package models

import (
  "bytes"
  "strings"
  "net/mail"
  "net/url"
  "fmt"
  "path"
  "net/http"
  "encoding/json"
  "text/template"
  "encoding/base64"
  "io/ioutil"
)
type ApiResponse struct {
  Code int    `json:"code"`
  Data string `json:"data"`
  Msg  string `json:"msg"`
}

// TemplateContext is an interface that allows both campaigns and email
// requests to have a PhishingTemplateContext generated for them.
type TemplateContext interface {
  getFromAddress() string
  getBaseURL() string
}

// PhishingTemplateContext is the context that is sent to any template, such
// as the email or landing page content.
type PhishingTemplateContext struct {
  From        string
  URL         string
  Tracker     string
  TrackingURL string
  EURL  string
  RId         string
  BaseURL     string
  BaseRecipient
}

// NewPhishingTemplateContext returns a populated PhishingTemplateContext,
// parsing the correct fields from the provided TemplateContext and recipient.
func NewPhishingTemplateContext(ctx TemplateContext, r BaseRecipient, rid string) (PhishingTemplateContext, error) {
  f, err := mail.ParseAddress(ctx.getFromAddress())
  if err != nil {
    return PhishingTemplateContext{}, err
  }
  fn := f.Name
  if fn == "" {
    fn = f.Address
  }
  templateURL, err := ExecuteTemplate(ctx.getBaseURL(), r)
  if err != nil {
    return PhishingTemplateContext{}, err
  }

  // For the base URL, we'll reset the the path and the query
  // This will create a URL in the form of http://example.com
  baseURL, err := url.Parse(templateURL)
  if err != nil {
    return PhishingTemplateContext{}, err
  }
  baseURL.Path = ""
  baseURL.RawQuery = ""

  phishURL, _ := url.Parse(templateURL)
  
  q := phishURL.Query()
  q.Set(RecipientParameter, rid)
  phishURL.RawQuery = q.Encode()

  trackingURL, _ := url.Parse(templateURL)
  trackingURL.Path = path.Join(trackingURL.Path, "/track")
  trackingURL.RawQuery = q.Encode()
  

  ipMap := map[string]string{
    "x.x.x.x": "www.example.com",
  }

  newURL := replaceIPWithDomain(phishURL.String(), ipMap)  

  // If QR logo needed, plz replace the value of the parameter logoUrl 
  apiURL := "http://qrcode.hlcode.cn/beautify/style/create?bgColor=%23FFFFFF&bodyType=1&content=" + newURL + "&down=0&embedPosition=0&embedText=&embedTextColor=%23000000&embedTextSize=38&eyeInColor=%23000000&eyeOutColor=%23000000&eyeType=8&eyeUseFore=1&fontFamily=0&foreColor=%23000000&foreColorImage=&foreColorTwo=&foreType=0&frameColor=&gradientWay=0&level=H&logoShadow=0&logoShap=1&logoUrl=&margin=2&rotate=30&size=400&qrCodeId=0&format=1"
  
  imageURL, err := GetQRCodeImageURL(apiURL)
  base64DataURL, err := convertImageToBase64(imageURL)
  return PhishingTemplateContext{
    BaseRecipient: r,
    BaseURL:       baseURL.String(),
    URL:           phishURL.String(),
    TrackingURL:   trackingURL.String(),
    EURL:  base64DataURL,
    Tracker:       "<img alt='' style='display: none' src='" + trackingURL.String() + "'/>",
    From:          fn,
    RId:           rid,
  }, nil
}

func convertImageToBase64(url string) (string, error) {
 
  imageContent, err := fetchRemoteImage(url)
  if err != nil {
    return "", err
  }


  base64Image := base64.StdEncoding.EncodeToString(imageContent)


  base64DataURL := "data:image/png;base64," + base64Image

  return base64DataURL, nil
}

func fetchRemoteImage(url string) ([]byte, error) {

  response, err := http.Get(url)
  if err != nil {
    return nil, err
  }
  defer response.Body.Close()


  imageContent, err := ioutil.ReadAll(response.Body)
  if err != nil {
    return nil, err
  }

  return imageContent, nil
}

func replaceIPWithDomain(url string, ipMap map[string]string) string {
  for ip, domain := range ipMap {

    url = strings.Replace(url, ip, domain, -1)
  }
  return url
}

func GetQRCodeImageURL(apiURL string) (string, error) {

  response, err := http.Get(apiURL)
  if err != nil {
    return "", fmt.Errorf("Error making GET request: %v", err)
  }
  defer response.Body.Close()


  if response.StatusCode < 200 || response.StatusCode >= 300 {
    return "", fmt.Errorf("Error: Unexpected status code %s", response.Status)
  }


  var apiResponse ApiResponse
  err = json.NewDecoder(response.Body).Decode(&apiResponse)
  if err != nil {
    return "", fmt.Errorf("Error decoding JSON: %v", err)
  }


  if apiResponse.Code != 0 {
    return "", fmt.Errorf("API returned an error: %s", apiResponse.Msg)
  }


  return apiResponse.Data, nil
}

// ExecuteTemplate creates a templated string based on the provided
// template body and data.
func ExecuteTemplate(text string, data interface{}) (string, error) {
  buff := bytes.Buffer{}
  tmpl, err := template.New("template").Parse(text)
  if err != nil {
    return buff.String(), err
  }
  err = tmpl.Execute(&buff, data)
  return buff.String(), err
}

// ValidationContext is used for validating templates and pages
type ValidationContext struct {
  FromAddress string
  BaseURL     string
}

func (vc ValidationContext) getFromAddress() string {
  return vc.FromAddress
}

func (vc ValidationContext) getBaseURL() string {
  return vc.BaseURL
}

// ValidateTemplate ensures that the provided text in the page or template
// uses the supported template variables correctly.
func ValidateTemplate(text string) error {
  vc := ValidationContext{
    FromAddress: "foo@bar.com",
    BaseURL:     "http://example.com",
  }
  td := Result{
    BaseRecipient: BaseRecipient{
      Email:     "foo@bar.com",
      FirstName: "Foo",
      LastName:  "Bar",
      Position:  "Test",
    },
    RId: "123456",
  }
  ptx, err := NewPhishingTemplateContext(vc, td.BaseRecipient, td.RId)
  if err != nil {
    return err
  }
  _, err = ExecuteTemplate(text, ptx)
  if err != nil {
    return err
  }
  return nil
}
