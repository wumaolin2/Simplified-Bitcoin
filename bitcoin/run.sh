#!/bin/bash
rm blockchain
rm *.db

go build -o blockchain *.go
./blockchain createBlockChain 1GQFsPpy9T2JN7wN1gdYP1fY3temRSXd3t
