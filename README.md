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

- "cp" not working on windows
- Add optional "copy mode" opposed to standard "clone mode" to initialize a service: copies the directory instead of cloning again
- Add "initializing" status to a service, in order to allow the server to start asap
- Pruning branches does not work with embedded git client (prune is not supported)
- Improve session page requesting only logs and status
- Piped commands not working (e.g. `docker run ... | xargs -I %s echo "%i"` )  
- Improve manager design
- Use host for main target forward
- Additional forward rules
- Configuration persistence via embeddable database (badgerDB?)
- Add possibility to always watch one or more branches and provide an always available session
- Add possibility to manually trigger a fetch in a git service folder
- Configuration reload
- Configuration CRUD via UI