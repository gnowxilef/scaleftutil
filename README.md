# scaleftutil
Stuff to interact with ScaleFT api

Super crude right now, takes one option, pattern and a simple pattern to do substring match against scaleft hostnames.  It will delete all that match without prompting.

You must set these env variables:
```
SCALEFT_KEY="somekey"
SCALEFT_KEY_SECRET="somesecret"
SCALEFT_TEAM="someteam"
SCALEFT_PROJECT="someproject"
```
