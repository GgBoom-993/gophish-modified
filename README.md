![gophish logo](https://raw.github.com/gophish/gophish/master/static/images/gophish_purple.png)

Gophish
=======

![Build Status](https://github.com/gophish/gophish/workflows/CI/badge.svg) [![GoDoc](https://godoc.org/github.com/gophish/gophish?status.svg)](https://godoc.org/github.com/gophish/gophish)

Gophish: Open-Source Phishing Toolkit

[Gophish](https://getgophish.com) is an open-source phishing toolkit designed for businesses and penetration testers. It provides the ability to quickly and easily setup and execute phishing engagements and security awareness training.



## Gophish - modified

- **Removal of some features**

This project  modified by gophishi v0.12.1 and modifying some basic characteristic of gophish to make it less detectable by security devices. Since only very simple modifications were made, there is no guarantee that it is completely hidden.

- **Trackable QR code templates**

With the development of diverse email phishing exercises, we no longer only use the traditional spoofing methods of clicking on a link or downloading an attachment. More and more scenarios induce users to scan malicious QR codes. Based on gophish's template capability, new QR code templates that can track each user have been added, simply by adding placeholders in the right places \<img src={\{.EURL}}>

- **Personalization**

The other changes are my personal thoughts. You can refer to the details of the changes below and make your own modifications.



## Details of the modifications

**Replaces some specially recognizable strings.**
```
find . -type f -exec sed -i 's/X-Gophish-Contact/X-Contact/g' {} +
find . -type f -exec sed -i 's/X-Gophish-Signature/X-Signature/g' {} +
sed -i 's/const ServerName = "gophish"/const ServerName = "mailServer"/' config/config.go
```

**Replaced the name of the parameter used for tracing**
rid -> ac
```
sed -i 's/const RecipientParameter = "rid"/const RecipientParameter = "ac"/g' models/campaign.go
```

**Added trackable QR code template {{.EURL}}**

If QR logo needed, Please change the value of the logoUrl parameter in line 84.
```
modified file: ./models/template_context.go
```

**Added function to customize request headers**
```
modified file: ./controllers/phish.go
```

**Change the port of the management panel and allow access from any address.**
./config.json
```
{
	"admin_server": {
		"listen_url": "0.0.0.0:8333",
		…………
	}
}
```

## Acknowledgements and References

I referenced the following articles for my revisions and thank the authors for their selfless sharing.

https://www.sprocketsecurity.com/resources/never-had-a-bad-day-phishing-how-to-set-up-gophish-to-evade-security-controls

https://mp.weixin.qq.com/s/EjrInb7bVMZP2ZOtNXUY5w
