#!/bin/bash

echo -n $(git describe --tags --dirty --always) > version.txt
