![gophish logo](https://raw.github.com/gophish/gophish/master/static/images/gophish_purple.png)

Gophish
=======

![Build Status](https://github.com/gophish/gophish/workflows/CI/badge.svg) [![GoDoc](https://godoc.org/github.com/gophish/gophish?status.svg)](https://godoc.org/github.com/gophish/gophish)

Gophish: Open-Source Phishing Toolkit

[Gophish](https://getgophish.com) is an open-source phishing toolkit designed for businesses and penetration testers. It provides the ability to quickly and easily setup and execute phishing engagements and security awareness training.

## Contents of the modification

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

**Added trackable QR code template , {{.EURL}}**
```
modified file: ./models/template_context.go
```

**Added function to customize request headers**
```
modified file: ./controllers/phish.go
```
