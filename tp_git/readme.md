# Tp Git

Simple temporary git repo cloner. I use this while doing code reviews, so I don't have to stash what I am working on then
switch branches. 

There are three different ways to use this command

single arg: 
`tp http://git.repo.com` clones given repo to temp dir and returns the dir 

two arg: 
`tp http://git.repo.com integration` clones given repo and branch to temp dir and returns the dir 

three arg: 
`tp http://git.repo.com integration code` clones given repo and branch to temp dir and returns the dir then opens the dir using the third arg command
Will clone repo and branch then open with third arg program. use `none`  to not open the repo in vscode. 