# Packaging a Node.js app for Shipyard

Zip the contents of the root directory of the app—do not include the root directory itself in the zip. You do not have to include node_modules; however, be sure that there is a properly configured package.json file in the base directory that can start your Node.js main file.

## Example

This directory contains a properly configured `package.json` and a single Node.js file, `index.js` needed to start the example appplication.

From this directory, zip the source like so:

```sh
# zip it so that the source file used by the start script
# and the package.json are in the root directory of the zip
zip -r example-app.zip index.js package.json
```

The result of executing `unzip -l example-app.zip` should be something like this:
```
Archive:  example-app.zip
  Length     Date   Time    Name
 --------    ----   ----    ----
     1053  04-14-16 13:07   index.js
      306  04-14-16 13:07   package.json
 --------                   -------
     1359                   2 files
```