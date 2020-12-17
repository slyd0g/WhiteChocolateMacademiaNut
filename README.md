# WhiteChocolateMacademiaNut

![WhiteChocolateMacademiaNut](./example.png)

## Description
Interacts with [Chromium-based browsers' debug port](https://blog.chromium.org/2011/05/remote-debugging-with-chrome-developer.html) to view open tabs, installed extensions, and cookies. Tested against Google Chrome and Microsoft Edge.

## Blogpost
- https://posts.specterops.io/hands-in-the-cookie-jar-dumping-cookies-with-chromiums-remote-debugger-port-34c4f468844e

## Usage
- Dump the user's open tabs and installed extensions
    - ```./WhiteChocolateMacademiaNut -p 4200 -d pages```
- Dump the user's cookies in human-readable format
    - ```./WhiteChocolateMacademiaNut --port 1337 --dump cookies --format human```
- Dump the user's cookies in raw JSON as returned by Chromium
    - ```./WhiteChocolateMacademiaNut --port 1234 --dump cookies --format raw```
- Dump the user's cookies in JSON with the name, value, domain, path, and modified expirationDate attribute to 10 years in the future (compatible with [Cookiebro extension](https://nodetics.com/cookiebro/))
    - ```./WhiteChocolateMacademiaNut -p 666 -d cookies -f modified```
- Dump the user's cookies in JSON if the cookie name or domain field contains `github` and modifies the expirationDate attribute
    - ```./WhiteChocolateMacademiaNut -p 4321 -d cookies -f modified -g github```
- Dump the user's cookies in human-readable format if the page title or url contains `facebook`
    -  ```./WhiteChocolateMacademiaNut --port 31415 --dump pages --grep facebook```

## References
- https://mango.pdf.zone/stealing-chrome-cookies-without-a-password
- https://github.com/defaultnamehere/cookie_crimes
