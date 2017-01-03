#!/bin/bash
set -e

rm -R '/media/ryan/Backup/GitHub/'
git clone --mirror 'https://github.com/rjoleary/dotfiles.git'             '/media/ryan/Backup/GitHub/dotfiles'
git clone --mirror 'https://github.com/rjoleary/templates.git'            '/media/ryan/Backup/GitHub/templates'
git clone --mirror 'https://github.com/rjoleary/programming-problems.git' '/media/ryan/Backup/GitHub/programming-problems'
git clone --mirror 'https://github.com/rjoleary/fractal.git'              '/media/ryan/Backup/GitHub/fractal'
git clone --mirror 'https://github.com/rjoleary/mutation-machine'         '/media/ryan/Backup/GitHub/mutation-machine'
git clone --mirror 'https://github.com/rjoleary/game-of-life'             '/media/ryan/Backup/GitHub/game-of-life'
