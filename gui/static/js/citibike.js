(function() {

  function checkBlockBeforeProgress(req, cat) {
    var required = req;
    var category = cat;

    var currentBlocks = JSON.parse($.ajax({
      url: '/blocks',
      type: 'GET',
          async: false // required before UI stream starts
        }).responseText);

    if (category == "type") {
      $.each(currentBlocks, function(k, v) {
        if (this.Type == required) {
          hopscotch.nextStep();
          return false;
        }
      });
    } else if (category == "endpoint") {
      $.each(currentBlocks, function(k, v) {
        if (this.Rule.Endpoint == required) {
          hopscotch.nextStep();
          return false;
        }
      });
    } else if (category == "interval") {
      $.each(currentBlocks, function(k, v) {
        if (this.Rule.Interval == required) {
          hopscotch.nextStep();
          return false;
        }
      });
    } else if (category == "map") {
      $.each(currentBlocks, function(k, v) {
        if (this.Type == "map") {
          if (this.Rule.Map.url == required) {
            hopscotch.nextStep();
            return false;
          }
        }
      });
    } else if (category == "path") {
      $.each(currentBlocks, function(k, v) {
        if (this.Type == "gethttp" || this.Type == "unpack") {
          if (this.Rule.Path == required) {
            hopscotch.nextStep();
            return false;
          }
        }
      });
    } else if (category == "filter") {
      $.each(currentBlocks, function(k, v) {
        if (this.Type == "filter") {
          if (this.Rule.Filter == required) {
            hopscotch.nextStep();
            return false;
          }
        }
      });
    }
    return false;
  }

  function checkConnectionsBeforeProgress(bF, bT) {
    var currentConnections = JSON.parse($.ajax({
      url: '/connections',
      type: 'GET',
      async: false // required before UI stream starts
    }).responseText);

    if (currentConnections.length == 0) {
      return false;
    }

    var blockFrom = bF;
    var blockTo = bT;

    var idFrom;
    var idTo;

    var currentBlocks = JSON.parse($.ajax({
      url: '/blocks',
      type: 'GET',
      async: false // required before UI stream starts
    }).responseText);

    $.each(currentBlocks, function(k, v) {
      if (this.Type == blockFrom) {
        idFrom = this.Id;
      }
      if (this.Type == blockTo) {
        idTo = this.Id;
      }
    });

    $.each(currentConnections, function(key, val) {
      if (this.FromId == idFrom && this.ToId == idTo) {
        hopscotch.nextStep();
        return false;
      }
    });
  }

  var tour = {
    id: "citibike",
    bubbleWidth: 350,
    showNextButton: false,
    showPrevButton: true,
    smoothScroll: false,
    
    steps: [
    {
      content: "<p>Welcome to streamtools.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
      xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        hopscotch.nextStep();
      }
    },

    {
      content: "<p>In this tutorial, we'll use streamtools to get live data on the availability of Citibikes at a particular station in NYC--namely, the station outside the NYT headquarters in Midtown Manhattan.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
      xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        hopscotch.nextStep();
      }

    },

    {
      content: "<p>To make sure our data is always up-to-date, we need some way to emit a message on a regular interval. We can do this with <b>emitter</b> blocks--in this case, a <span class=\"tutorial-blockname\">ticker</span> block.</p><p>Click the hamburger button to see a list of every block in streamtools.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        hopscotch.nextStep();
      }
    },

    {
      content: "<p>Click <span class=\"tutorial-blockname\">ticker</span> to add that block.</p><p>You can click and drag blocks to move them around on screen. You can delete a block by clicking it (to select it) and pressing the Delete key.</p><p>Click Next when you're ready.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkBlockBeforeProgress("ticker", "type");
      }
    },

    {
      content: "<p>Most blocks have <b>rules</b>, which define that block's behavior. You can double-click a block to edit its rules.</p><p>Double-click your <span class=\"tutorial-blockname\">ticker</span> block.</p><p>Let's set our interval to 10 seconds. Type <span class=\"tutorial-url\">10s</span> into the Interval box and click Update.</p><p>After that, click Next.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkBlockBeforeProgress("10s", "interval");
      }
    },

    {
      content: "<p>Before we can start making GET requests, we need to specify the URL from which we are getting the data.</p><p>We will use a <span class='tutorial-blockname'>map</span> block for this, mapping the key 'url' to our URL. This is a <b>flow block</b>: one that transforms or manipulates the stream.</p><p>Double-click anywhere on screen to add a block.</p><p>Type in <span class='tutorial-blockname'>map</span> and hit Enter.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkBlockBeforeProgress("map", "type");
      }
    },

    {
      content: "<p>Double-click the <span class='tutorial-blockname'>map</span> block to edit its parameters.</p> <p>The <span class='tutorial-blockname'>map</span> block takes a <a href='https://github.com/nytlabs/gojee' target='_new'>gojee</a> expression.</p> <p>Our map will look like this:<br/><span class='tutorial-url'>{ \"url\": \"\'http://citibikenyc.com/stations/json\'\" }</span></p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkBlockBeforeProgress("\'http://citibikenyc.com/stations/json\'", "map");
      }
    },

    {
      content: "<p>Each block also has a set of <b>routes</b>, which can receive data, emit data, or respond to queries. Routes can be connected, allowing data to flow between blocks.</p><p>Let's connect the two, so every 10s, we map this URL.</p><p>Click the OUT box on your <span class=\"tutorial-blockname\">ticker</span> box (the bottom black box).</p><p>Connect it to the IN on your <span class=\"tutorial-blockname\">map</span> (the top black box).</p> ",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkConnectionsBeforeProgress("ticker", "map");
      }
    },

    {
      content: "<p>Now we need to actually get our data. We'll make this GET request with a <span class=\"tutorial-blockname\">gethttp</span> block.</p><p>Double-click anywhere on screen to add a block.</p><p>Type in <span class=\"tutorial-blockname\">gethttp</span> and hit Enter.</p> ",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkBlockBeforeProgress("gethttp", "type");
      }
    },

    {
      content: "<p>Double-click on your <span class=\"tutorial-blockname\">gethttp</span> block to edit it.</p><p>Remember how we mapped our URL? In <a href=\"https://github.com/nytlabs/gojee\" target=\"_new\">gojee syntax</a>, our \"url\" key becomes the path <span class=\"tutorial-url\">.url</span>.</p><p>Put that in the Path parameter, click Update, then click Next.</p> ",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkBlockBeforeProgress(".url", "path");
      }
    },

    {
      content: "<p>Now let's connect our <span class=\"tutorial-blockname\">map</span> block to our <span class=\"tutorial-blockname\">gethttp</span> block.</p><p>That way, we'll make a GET request to that URL every 10s.</p><p>Click the OUT box on your <span class=\"tutorial-blockname\">map</span> box (the bottom black box).</p><p>Connect it to the IN on your <span class=\"tutorial-blockname\">gethttp</span> (the top black box).</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkConnectionsBeforeProgress("map", "gethttp");
      }
    },

    {
      content: "<p>If you view the <a href=\"http://citibikenyc.com/stations/json\" target=\"_new\">JSON data</a> in your browser, you'll see that all the data is in a big array.</p><p>The key wrapping up all the data about individual stations is <span class=\"tutorial-url\">stationBeanList</span>.</p><p>In order to be able to manipulate and filter this data, we need to unpack it first.</p><p>That\'s where the <span class=\"tutorial-blockname\">unpack</span> block comes in handy: it takes an array of objects and emits each object as a separate message. Double-click and create it anywhere on-screen.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkBlockBeforeProgress("unpack", "type");
      }
    },

    {
      content: "<p>Double-click on the <span class=\"tutorial-blockname\">unpack</span> block to edit its rule.</p><p>Set its Path to <span class=\"tutorial-url\">.stationBeanList</span> and click Next.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkBlockBeforeProgress(".stationBeanList", "path");
      }
    },

    {
      content: "<p>Now let's connect our <span class=\"tutorial-blockname\">gethttp</span> (the thing giving us the JSON) to our <span class=\"tutorial-blockname\">unpack</span> (the thing iterating over that JSON).</p><p>Connect the two and click Next.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkConnectionsBeforeProgress("gethttp", "unpack");
      }
    },

    {
      content: "<p>Right now we're getting data about every station. Let's filter out every station other than the one outside the NYT headquarters.</p><p>For this, we'll use a <span class=\"tutorial-blockname\">filter</span> block.</p><p>Click Next once you've made it.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkBlockBeforeProgress("filter", "type");
      }
    },

    {
      content: "<p>The <span class=\"tutorial-blockname\">filter</span> block contains a rule, and emits/discards incoming messages based on whether the rule evaluates to true.</p><p>The station nearest the NYT HQ is <span class=\"tutorial-url\">\'W 41st St & 8 Ave\'</span>.</p><p>Our <span class=\"tutorial-blockname\">filter</span> rule will look like this:</p><p><span class=\"tutorial-url\">.stationName == \'W 41 St & 8 Ave\'</span></p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkBlockBeforeProgress(".stationName == 'W 41 St & 8 Ave'", "filter");
      }
    },

    {
      content: "<p>Connect your <span class=\"tutorial-blockname\">unpack</span> and <span class=\"tutorial-blockname\">filter</span> blocks.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkConnectionsBeforeProgress("unpack", "filter");
      }
    },

    {
      content: "<p>Usually, we'll want to send our data to an external system, such as a file or our console. <b>Sink blocks</b> allow us to do that. In our case, we'll send our data to the log.</p><p>The <span class=\"tutorial-blockname\">tolog</span> block logs your data to the console and the log built into streamtools.</p><p>Add it and click Next.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkBlockBeforeProgress("tolog", "type");
      }
    },

    {
      content: "<p>Finally, connect your <span class=\"tutorial-blockname\">filter</span> and <span class=\"tutorial-blockname\">tolog</span> blocks.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkConnectionsBeforeProgress("filter", "tolog");
      }
    },

    {
      content: "<p>Now, every 10s, your log will be updated with your newest filtered live data.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
      xOffset: "center",
      showCTAButton: false,
      showNextButton: true,
    },

    {
      content: "<p>Now that you've finished this example, try building on it by incorporating more of streamtools' features.</p><p>Check out <a href=\"http://nytlabs.github.io/streamtools\" target=\"_new\">detailed block descriptions</a>, or play with some <a href=\"http://nytlabs.github.io/streamtools/demos/\">pre-built demos</a>.",
      target: "#log",
      placement: "top",
      yOffset: -20,
      xOffset: "center",
      showCTAButton: false,
      showNextButton: true,
    },

    ]
  };

// Start the tour!
hopscotch.startTour(tour, 0);


})();
