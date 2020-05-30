# api-server
###### [![Build Status](https://travis-ci.org/MyHomeworkSpace/api-server.svg?branch=master)](https://travis-ci.org/MyHomeworkSpace/api-server) [![Created by the MyHomeworkSpace Team](https://img.shields.io/badge/Created%20by-MyHomeworkSpace%20Team-3698dc.svg)](https://github.com/MyHomeworkSpace)
The MyHomeworkSpace API server. See also the [API reference](https://support.myhomework.space/apireference) and [Web client](https://github.com/MyHomeworkSpace/client).

## Requirements
* A modern version of Go (1.11 or newer, recommended to use whatever the latest version is)
* MySQL (preferably version 8, but 5.6 or 5.7 work too)
* Redis
* MailHog (only required if you want to test emails)
* OpenResty (or any other web server, but the guide assumes OpenResty)
* Roamer

Go should be downloaded from [the official website](https://golang.org/dl/).

MySQL can be installed via your operating system's package manager, or via [the official site](https://dev.mysql.com/downloads/mysql/) if on Windows or macOS. You might also want some tool to connect to and manage your MySQL database, such as [MySQL Workbench](https://dev.mysql.com/downloads/workbench/).

Redis should be installed via your operating system's package manager, but you can also [compile the latest version from their website](https://redis.io/) if you want.

MailHog can be downloaded from their [releases page](https://github.com/mailhog/MailHog/releases/v1.0.0). Windows users can download the appropriate .exe file and run it. Linux or macOS users should download either the `_linux_amd64` or `_darwin_amd64` executable. Then, open a terminal, `cd` to the location of the executable, `chmod +x <name of executable>`, and `./<name of executable>`. That last command starts MailHog with a web UI at http://localhost:8025. Later on, you'll only need to run that last command to start MailHog.

OpenResty can be installed by following [their instructions](https://openresty.org/en/download.html). You can use a different server, such as nginx, Apache, or Microsoft IIS, but you might have to alter the instructions below.

Roamer can be installed by following [their instructions](https://github.com/thatoddmailbox/roamer/wiki/installation).

## Setting up (for development)
This guide might seem long and complicated! However, don't worry: it's supposed to be somewhat methodical and detailed, which is why there's so many steps. You'll be up and running in no time!

1. First, you'll need to create a new database in MySQL. You can call it something clever, like `myhomeworkspace`.
2. You'll also want to create a new user account in MySQL that has full access to the `myhomeworkspace` database. Make its password something random and secure. You'll need it later.
3. Run `go get github.com/MyHomeworkSpace/api-server`. This will download this repository (and its dependencies) into the appropriate location of your $GOPATH.
4. In a terminal, change directory into the location where the code was downloaded (on Linux/macOS, this will be something like `~/go/src/github.com/MyHomeworkSpace/api-server`)
5. Run `roamer setup`. Roamer will create a `roamer.local.toml` file.
6. Open the `roamer.local.toml` file in a text editor of your choice. Update the line that begins with `DSN=` to refer to the database and user that you just created. It should look something like `DSN = "myhomeworkspace:mySuperSecurePassword@tcp(localhost:3306)/myhomeworkspace"`.
6. Run `go run github.com/MyHomeworkSpace/api-server`. It will display an error, this is normal.
7. A new `config.toml` file will have appeared in the api-server folder. Open it in a text editor.
8. This part can get somewhat specific to your setup. We _are_ making some assumptions, but these assumptions will be valid if you follow this guide.
9. In the `[server]` section: since the default port, 3000, is somewhat common, we recommend changing it. This guide uses port 3001.
10. In the `[database]` section: set `Password` to the password you set for your MySQL account in step 2.
11. For the `[email]` section (assuming you're using MailHog):
```
Enabled = true
FromAddress = "hello@myhomework.invalid"
FromDisplay = "MyHomeworkSpace <hello@myhomework.invalid>"
SMTPHost = "localhost"
SMTPPort = 1025
SMTPSecure = false
SMTPUsername = "hello@myhomework.invalid"
SMTPPassword = "password123"
```
12. For the `[redis]` section: the defaults are fine, unless you've changed Redis' default settings.
13. For the `[cors]` section: set `Enabled` to true.
14. You can leave the `[errorlog]`, `[feedback]`, `[tasks.slack]`, and `[mit]` sections as-is.
15. Save your changes to `config.toml`, and re-run `go run/github.com/MyHomeworkSpace/api-server` in the terminal from step 6. Leave this running.
16. Visit http://localhost:3001 in your browser. You should see a page saying "MyHomeworkSpace API server".
17. Now, it's time to set up OpenResty. OpenResty will act as the main webserver that everything goes through. That way, you're able to host both the API server and the client on the same computer, or even other unrelated projects!
18. First, however, we'll set up local domain records. This is the myhomework.invalid that was mentioned earlier, and will make it so that you can access your local copy of MyHomeworkSpace by going to myhomework.invalid in your browser.
19. You need to open the _hosts_ file on your computer. On Linux and macOS, this is located at `/etc/hosts`. On Windows, this is located at `C:\Windows\System32\drivers\etc\hosts`. Editing this file will likely require administrator privileges.
20. In the hosts file, add the following entries:
```
127.0.0.1	myhomework.invalid
127.0.0.1	api-v2.myhomework.invalid
127.0.0.1	app.myhomework.invalid
```
21. Save your changes, and try to go to http://api-v2.myhomework.invalid in your browser. You should see a generic "Welcome to OpenResty!" page.
22. Now, open up the OpenResty configuration file. On Linux, this is probably at `/etc/openresty/nginx.conf`. On macOS, this is probably at `/usr/local/etc/openresty/nginx.conf`. You might need administrator privileges to edit this file.
23. You'll see a big `http {}` block, with lots of stuff in it. At the bottom of this block, *right before the last `}`*, you will want to add the following (which will set up the API server, client, and main website all at once):
```
	map $http_upgrade $mhs_connection_upgrade {
		default upgrade;
		''      close;
	}

	server {
		listen 80;
		listen [::]:80;

		server_name api-v2.myhomework.invalid;

		location / {
			proxy_pass http://127.0.0.1:3001;
		}
	}

	server {
		listen 80;
		listen [::]:80;

		server_name app.myhomework.invalid;

		proxy_buffering off;

		keepalive_disable safari;
		keepalive_timeout 0;

		location / {
			proxy_pass http://127.0.0.1:9001;

			proxy_http_version 1.1;
			proxy_set_header Upgrade $http_upgrade;
			proxy_set_header Connection $mhs_connection_upgrade;
		}
	}

	server {
		listen 80;
		listen [::]:80;

		server_name myhomework.invalid;

		location / {
			proxy_pass http://127.0.0.1:4003;
		}
	}
```
24. Save these changes. Now you need to tell OpenResty that you changed the config file. This can be done by running `sudo openresty -s reload`. You can also do `sudo openresty -t` to verify the syntax of the config file.
25. Try going to http://api-v2.myhomework.invalid in your browser now. You should see the "MyHomeworkSpace API server" page from before.
26. Congratulations! You've set up the MyHomeworkSpace API server. You probably want to set up the [client](https://github.com/MyHomeworkSpace/client) or [website](https://github.com/MyHomeworkSpace/website) now.

## Running the server
Once you've done the setup above, running the server is as easy as:
```
cd ~/go/src/github.com/MyHomeworkSpace/api-server
go run github.com/MyHomeworkSpace/api-server
```
If you're testing email-related things, you should also make sure MailHog is set up and running. See the Requirements section if you need help with that.
