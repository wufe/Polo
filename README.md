# Polo

## Proxy your application history

Polo allows you to create a web server which provides the ability to serve your web application in a specific time in history, using git.  

You just need to specify the git commit or the branch. A new session will be created for you to navigate into.

Although it provides HTTPS support, it is not intended to be used in production.  

***

## Getting started

- Download Polo from the release page
- Create one or more configuration files for your application
- Start Polo

***

## Configuration

You must provide at least one yaml configuration file describing your service remote and how to build and run it.  
The configuration file must be put next to the Polo executable file.  
You can find an example of a configuration file with all the options in the folder *examples*.  

***

## Known issues / missing features

- Fix service sessions not updating in main manager page  
- Add "last author" and "last update" for each branch  in main manager page
- Improve session page requesting only logs and status
- Piped commands not working (e.g. `docker run ... | xargs -I %s echo "%i"` )  
- Improve manager design
- Use host for main target forward
- Additional forward rules
- Configuration persistence via sqlite
- Configuration reload
- Configuration CRUD via UI