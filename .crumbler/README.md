i have to refresh to see agent responses in the channel but i should not need to do that

rip out the sandbox feature this will run in a sandbox that is one level above this we are not sandboxing at this level rip out allowed commands thats not going to be a thing

the agent will have a current project that its working on and that'll determine the agent's shell pwd

the whole point is that the projects are on the agent machine that the server is running on

project erd prd and decisions are completely wrong it works like this there is a git project going on right in there there are 3 folders erd, prd and decisions these folders contain markdown documents for each of those things there isn't anything in particular in the db about it and state should be dictated by what is physically on disk critically each of these are simply files on disk and should not have anything in the db backing them up

it is expected that agent is going to be a coding agent working alone or in parallel on projects within the projects directory

better markdown support for multiline codeblocks in the chat interface

typing... indicia support

memory is much too aggressive and i can't stand the way its done rip it out its going to work like this now

- agent will be told that memories are stored in the chat history and can be searched out of there for keywords
- memories will also be stored in the project folder in ./memories/yyyy-mm-dd-hh-mm-title-in-kebab.md docs
- use grep to look through these memories in the project dir

agent needs default prompts that'll help it more we can't lean on the user to do it, here are some built in persona templates we need its all around building software all of these prompts need to be opinionated

backend is go frontend is vite+react+ts spa

- code architect
  - defines interfaces and documentation
  - api design
  - db design
  - seams and code organization

- senior backend engineer

- senior frontend engineer

- senior qa engineer

- product manager
