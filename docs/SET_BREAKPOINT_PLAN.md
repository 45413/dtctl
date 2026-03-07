# POC start

We're gonna implement support for setting a breakpoint using the Live Debugger.
The goal is to have a POC.
The minimal requirements for a POC is to at least:
1. Get the current workspace the user has, by using getOrCreateWorkspace GraphQL (or maybe getOrCreateWorkspaceV2, you need to check). We'll use the clientName parameter and set it with "dtctl".
2. Update the workspace's filters in order to choose the servers we want to debug. This will be done by updateWorkspaceV2 graphql.
3. Set a breakpoint by setting the filename and line number. This will be done with CreateRuleV2 graphql.

So, here are example commands:
if I want to debug all the servers from the k8s.namespace.name=prod I would start with:

dtctl debug --filters k8s.namespace.name=prod

I should be able to set multiple filters seperated with a comma.

And then I should be able to place a breakpoint by:

dtctl debug --breakpoint OrderController.java:306

This will place a breakpoint on file OrderController.java line 306.

You should show the return output of the graphql requests, think about how you should do that.

You can see all the details of the graphql and code usage in the folder devobs-vs-code-plugin where I cloned a repository from a different project that does all these things. Take inspiration and learn from it.
