#!/usr/bin/env pwsh
# Quick wrapper to run npm with Node.js from N:\Apps\node
$env:PATH = "N:\Apps\node;$env:PATH"
& "N:\Apps\node\npm.cmd" @args
