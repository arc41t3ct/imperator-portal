# Imperator Portal

<div align="center">
    <img src="docs/images/logo.jpg" style="width:200px" />
</div>

## Infrstructure

To develop quickly local there is another public repo that is useful:

[Imperator Dev Infra](https://github.com/arc41t3ct/imperator-dev-infra)

## Dotenv .env

The Imperator Framework requires all these envrionment variables below. Make sure to add this .evn file 
to your root before you build it. It is in the .gitignore on purpose so you do not ship your credentials. 
Have this file in your pipeline to build your app with real credentials coming from a credential store 
replacing passwords and other secrets in this file.

```
# APP Configuration /.env
# the name of the app
APP_NAME=Imperator
APP_URL="http://localhost:4000"

# DEBUG Configuration 
# app running in debug mode - this will put jet templates in development 
# mode for easy page reresh and causes extra logging (DEBUG=true) is like Dev mode
# (DEBUG=false) is like Prod mode
DEBUG=true
# DEBUG=false

# RENDERER Configuration 
# which template engine would you like to use? jet or go
#RENDERER=go
RENDERER=jet

# should we use https
SSL_ENABLED=false

# PORT Configuration 
# to run the application on
PORT=4000

# SERVER configuration
# www.excample.com or other
SERVER_NAME=localhost

# SESSION Configuration 
# store: cookie, redis, mysql, postgres
# SESSION_TYPE=cookie
SESSION_TYPE=redis

# SESSION_TYPE=postgres
# COOKIE Configuration
#change this to your site name or so
COOKIE_NAME=imperator 
# minutes
COOKIE_LIFETIME=1440 
COOKIE_PERSISTS=true
# encrypt todo
COOKIE_SECURE=false 
#should be domain site is running on
COOKIE_DOMAIN=localhost 

# DATABASE Configuration
# can also be mysql
DATABASE_TYPE=postgres 
#DATABASE_TYPE=mysql
DATABASE_HOST=localhost
DATABASE_PORT=5432
#DATABASE_PORT=3306
DATABASE_USER=imperator
DATABASE_PASSWORD=password
DATABASE_NAME=imperator
DATABASE_SSL_MODE=disable

# REDIS Configuration
REDIS_HOST="localhost:6379"
REDIS_PASSWORD=password
REDIS_PREFIX=imperator

# CACHE Configuration
# currently only redis is supported
# CACHE_TYPE=badger
CACHE_TYPE=redis

# ENCRYPTION Configuration
# generated with ./imperitor make key
ENCRYPTION_KEY=cukSX7a8aa97SAX6_as766-asc1229SS

# SMTP Configuration
SMTP_HOST=localhost
SMTP_USERNAME=
SMTP_PASSWORD=
# port 1025 used by the mailhog tool to test emails
SMTP_PORT=1025
SMTP_ENCRYPTION=none

# MAIL Configuration
# these should be matching what is setup in mailgun
MAIL_DOMAIN=mg.awesoome-site.store
MAIL_FROM_NAME="Awesome.Site Team"
MAIL_FROM_ADDRESS="customer.service@awsome-site.store"

# MAILER Configuration
# setting MAILER_API=smtp allows you to use internal SMTP setting
# but we can turn on mailgun again by commenting it out
MAILER_API=smtp
# MAILER_API=mailgun
MAILER_KEY=some-mail-gun-key-for-sending
MAILER_URL=https://api.mailgun.net

```
