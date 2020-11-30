#!/bin/bash

cp .devcontainer/config.toml config.toml
cp .devcontainer/roamer.local.toml roamer.local.toml
roamer upgrade
