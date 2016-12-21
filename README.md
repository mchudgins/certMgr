# Certificate Manager

Let's Encrypt is awesome.  However, if you need to have ten's (or hundreds) of certificates for your domain,
then you'll likely run into the rate limits imposed by Let's Encrypt.  This project provides a simple mechanism
to create and retrieve self-signed certificates.

### ToDo
* use TLS
* use mutual auth
* fix prometheseus /metrics on frontend
* verify that openshift builder works
* add configserver, hystrix, turbine to openshift deployments
* swagger ui
