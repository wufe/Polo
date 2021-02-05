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

You must provide at least one yaml configuration file describing your application remote and how to build and run it.  
The configuration file must be put next to the Polo executable file.  
You can find an example of a configuration file with all the options in the folder *examples*.  

***

## Known issues / missing features

- Add support to command concatenations (; and &&)
- Update session-helper design
- Update session page design (fading terminal, better scrollbar, estimated time required, progress bar)
- Add possibility to start healthchecking early
- Ended sessions cleanup (folder structure)
- Add optional "copy mode" opposed to standard "clone mode" to initialize a application: copies the directory instead of cloning again
- Add "initializing" status to a application, in order to allow the server to start asap
- Pruning branches does not work with embedded git client (prune is not supported)
- Additional forward rules
- Configuration persistence via embeddable database (badgerDB?)
- Add possibility to always watch one or more branches and provide an always available session
- Add possibility to manually trigger a fetch in a git application folder
- Configuration reload
- Configuration reload via watching files
- Configuration CRUD via UI