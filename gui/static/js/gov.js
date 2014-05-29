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
    id: "gov",
    bubbleWidth: 350,
    showNextButton: false,
    smoothScroll: false,
    
    steps: [
    {
      content: "Welcome to streamtools.",
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
      content: "<p>In this tutorial, we'll use streamtools to get data on people clicking shortlinks to get to US government websites.</p>",
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
      content: "<p>First, we need a <b>source block</b>: that is, a block that hooks into another system and collects messages that are then emitted into streamtools.</p><p>The US government short-link API is a long-lived http stream, so the source block we'll use is the <span class=\"tutorial-blockname\">fromhttpstream</span> block.</p><p>Click the hamburger button to see a list of every block in streamtools.</p>",
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
      content: "<p>Click <span class=\"tutorial-blockname\">fromhttpstream</span> to add that block.</p><p>You can click and drag blocks to move them around on screen.</p><p>Click Next when you're ready.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkBlockBeforeProgress("fromhttpstream", "type");
      }
    },

    {
      content: "<p>Most blocks have <b>rules</b>, which define that block's behavior. You can double-click a block to edit its rules.</p><p>Double-click your block and type <span class=\"tutorial-url\">http://developer.usa.gov/1usagov</span> into the endpoint. Click Update, then click Next.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkBlockBeforeProgress("http://developer.usa.gov/1usagov", "endpoint");
      }
    },

    {
      content: "<p>Usually, we'll want to send our data to an external system, such as a file or our console. <b>Sink blocks</b> allow us to do that. In our case, we'll send our data to the log.</p><p>Double-click anywhere on screen to add a block.</p><p>Type in <span class=\"tutorial-blockname\">tolog</span> and hit Enter.</p>",
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
      content: "<p>Each block also has a set of <b>routes</b>, which can receive data, emit data, or respond to queries. Routes can be connected, allowing data to flow between blocks.</p><p>Let's connect our two blocks, so we have our data streaming into our log.</p><p>Click the OUT box on your <span class=\"tutorial-blockname\">fromhttpstream</span> box (the bottom black box).</p><p>Connect it to the IN on your <span class=\"tutorial-blockname\">tolog</span> (the top black box).</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
            xOffset: "center",
      showCTAButton: true,
      ctaLabel:"Next",
      onCTA: function() {
        checkConnectionsBeforeProgress("fromhttpstream", "tolog");
      }
    },

    {
      content: "<p>Now click the log (this black bar) to view your data.</p>",
      target: "#log",
      placement: "top",
      yOffset: -20,
      xOffset: "center",
      showCTAButton: false,
      showNextButton: true,
    },

    {
      content: "<p>Now that you've finished this simple example, try building on it by incorporating more of streamtools' features.</p><p>Check out <a href=\"http://nytlabs.github.io/streamtools\" target=\"_new\">detailed block descriptions</a>, or play with some <a href=\"http://nytlabs.github.io/streamtools/demos/\">pre-built demos</a>.",
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
