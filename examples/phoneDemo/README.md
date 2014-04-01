# phone demo

This demo uses the ability of HTML5 to sense the orientation of some smart
phones. The orientation is collected from a bunch of phones, POSTed to
streamtools and averaged. It's a little advanced, so you need to do a bit more
than usual to get everything working.  

To set up the demo you'll need to edit the vote.html file, and change the `IP`
variable to your streamtood demo IP. If you're running locally, google for
"what's my IP" or use `ifconfig` to find out your IP. 

To run it you'll need to run streamtools, import the `phoneDemo.json` pattern, and also run a
webserver to serve up the HTML. On linux or osx you can run

    python -m SimpleHTTPServer 8000

in this folder and then visit `http://yourIpAddress:8000/vote.html` on your phone to send data. 
Try getting a bunch of people to visit that same address to really see how this
works. To see the results of everyone using their phone to vote visit
`http://localhost:8000` on your local computer to see the results. 
