FROM php:5.6-apache
RUN a2enmod rewrite vhost_alias
RUN mkdir /phplib
COPY ./httpd.dynamic.conf /etc/apache2/sites-available/000-default.conf
