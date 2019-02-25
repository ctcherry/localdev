run:
	@sudo -n -- true 2>/dev/null || echo "Please enter your password so we can setup the custom DNS resolver"
	@echo "nameserver 127.0.0.1\nport 19053\norder 162\ntimeout 1" | sudo tee /etc/resolver/loc > /dev/null
	docker-compose up
	docker-compose down
	@sudo -n -- true 2>/dev/null || echo "Please enter your password so we can cleanup the custom DNS resolver"
	sudo rm /etc/resolver/loc

