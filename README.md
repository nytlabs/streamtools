streamtools
===========

[![Build Status](https://travis-ci.org/nytlabs/streamtools.png?branch=master)](https://travis-ci.org/nytlabs/streamtools)

Streamtools is a creative tool for working with streams of data. It provides a vocabulary of data processing operations, called blocks, that can be connected to create online data processing systems without the need for programming or complicated infrastructure. 

Streamtools is built upon a few core principles: 
- Working with data should be a responsive, exploratory practice. Streamtools allows you to immediately ask questions of the data as it flows through the system (see: [Creative Abduction](https://github.com/nytlabs/streamtools/wiki#philosophy)). 
- In the real world, the character of your data is constantly changing. We designed Streamtools not only to reflect how your data is changing but to let you work with that change (see: [Non-Stationarity](https://github.com/nytlabs/streamtools/wiki#philosophy)).  
- Working with data should not require complex engineering. Streamtools provides a visual interface and an expressive toolset for working with streams of data. 

Streamtools is an open source project written in Go and is intended to be used with streams of JSON.

![s3 polling
example](https://raw.github.com/mikedewar/streamtools/master/examples/crazy_example.png)

getting started
===============

1. Find a computer to play with. It needs to be Linux or OSX. 
2. Download the latest [release](https://github.com/nytlabs/streamtools/releases). You need `st-linux` if you're on linux or `st-darwin` if you're on osx.
3. In a terminal, change directory to wherever you downloaded the file. 
4. Run `chmod +x st-linux` if you're on linux or `chmod +x st-darwin` on osx. This makes the file you downloaded executeable. 
5. Now launch streamtools by typing `./st-linux` if you're on linux or `./st-darwin` if you're on osx. Your terminal should say `starting stream tools on port 7070`.
6. To find the UI visit [http://localhost:7070](http://localhost:7070) in a browser. If you're not running streamtools locally you need some way of accessing port 7070 on your remote box.
7. Go through our [Hello World](https://github.com/nytlabs/streamtools/wiki/Hello-world) pattern!
8. Look through the rest of our [patterns](https://github.com/nytlabs/streamtools/wiki#patterns) for inspiration and guidance. 

Good luck!

health warning
==============

*Note that streamtools is very new!* This means we're developing it very rapidly, and some things aren't going to work. If you find a bug please do let us know! And, if you think of something you'd like to see, please do request it! Both of these things can be done on our [issues page](https://github.com/nytlabs/streamtools/issues?milestone=&page=1&state=open). 

![screen shot 2014-02-06 at 4 22 51 pm](https://f.cloud.github.com/assets/597897/2103977/151b99ce-8f75-11e3-99b9-188024ce742a.png)

contributing
============
At the moment (Feb 14) we are in the midst of a large re-write! So stay tuned: you will see large changes in March. Contributing after the rewrite will be a lot easier!
