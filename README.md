localdev
--------

localdev is a set of docker containers managed by docker-compose designed to support the local development of websites on OSX.

You can create a folder in the VHOSTS_DIR (by default set to ~/sites, you can change this in the .env file) named `whateverdomain.loc` and put an index.html file in it, and immediately access it via HTTP and HTTPS.

It does this by running a dynamic virtual hosting setup based on apache, including a dynamic TLS proxy (based on tlself), and DNS server (dnsmasq) that points *.loc domains to localhost.

Platforms supported: OSX

Usage
-----

`make run`

You may need to enter your password to setup the custom resolver.
On the first run it will also give you a command you need to run that will trust the root cert, enabling dynamic HTTPS support.
