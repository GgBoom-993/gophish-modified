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
  
  //这里x.x.x.x换成和钓鱼邮箱域名绑定的ip地址
  ipMap := map[string]string{
    "x.x.x.x": "www.example.com",
  }
  // 调用替换函数
  newURL := replaceIPWithDomain(phishURL.String(), ipMap)  

  apiURL := "http://qrcode.hlcode.cn/beautify/style/create?bgColor=%23FFFFFF&bodyType=1&content=" + newURL + "&down=0&embedPosition=0&embedText=&embedTextColor=%23000000&embedTextSize=38&eyeInColor=%23000000&eyeOutColor=%23000000&eyeType=8&eyeUseFore=1&fontFamily=0&foreColor=%23000000&foreColorImage=&foreColorTwo=&foreType=0&frameColor=&gradientWay=0&level=H&logoShadow=0&logoShap=1&logoUrl=https:%2F%2Foss.hlcode.cn%2Fserver%2F2024%2F01%2F04%2F164890235818.png&margin=2&rotate=30&size=400&qrCodeId=0&format=1"
  
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
  // 获取远程图片的内容
  imageContent, err := fetchRemoteImage(url)
  if err != nil {
    return "", err
  }

  // 将图片内容转换为Base64编码
  base64Image := base64.StdEncoding.EncodeToString(imageContent)

  // 构造Base64格式的图片URL
  base64DataURL := "data:image/png;base64," + base64Image

  return base64DataURL, nil
}

func fetchRemoteImage(url string) ([]byte, error) {
  // 发起HTTP请求获取远程图片内容
  response, err := http.Get(url)
  if err != nil {
    return nil, err
  }
  defer response.Body.Close()

  // 读取图片内容
  imageContent, err := ioutil.ReadAll(response.Body)
  if err != nil {
    return nil, err
  }

  return imageContent, nil
}

func replaceIPWithDomain(url string, ipMap map[string]string) string {
  for ip, domain := range ipMap {
    // 将IP地址替换为域名
    url = strings.Replace(url, ip, domain, -1)
  }
  return url
}

func GetQRCodeImageURL(apiURL string) (string, error) {
  // 发送GET请求
  response, err := http.Get(apiURL)
  if err != nil {
    return "", fmt.Errorf("Error making GET request: %v", err)
  }
  defer response.Body.Close()

  // 检查响应是否成功 (状态码为 200-299)
  if response.StatusCode < 200 || response.StatusCode >= 300 {
    return "", fmt.Errorf("Error: Unexpected status code %s", response.Status)
  }

  // 解析JSON响应
  var apiResponse ApiResponse
  err = json.NewDecoder(response.Body).Decode(&apiResponse)
  if err != nil {
    return "", fmt.Errorf("Error decoding JSON: %v", err)
  }

  // 检查API响应中的错误码
  if apiResponse.Code != 0 {
    return "", fmt.Errorf("API returned an error: %s", apiResponse.Msg)
  }

  // 返回图片地址
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
