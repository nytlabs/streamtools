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
          return true;
        }
      });
    } else if (category == "endpoint") {
      $.each(currentBlocks, function(k, v) {
        console.log(this.Rule.Endpoint);
        if (this.Rule.Endpoint == required) {
          hopscotch.nextStep();
          return true;
        }
      });
    } else if (category == "interval") {
      $.each(currentBlocks, function(k, v) {
        console.log(this.Rule.Interval);
        if (this.Rule.Interval == required) {
          console.log("interval matches");
          hopscotch.nextStep();
          return true;
        }
      });
    } else if (category == "map") {
      $.each(currentBlocks, function(k, v) {
        if (this.Type == "map") {
          if (this.Rule.Map.url == required) {
            hopscotch.nextStep();
            return true;
          }
        }
      });
    } else if (category == "path") {
      $.each(currentBlocks, function(k, v) {
        if (this.Type == "gethttp" || this.Type == "unpack") {
          if (this.Rule.Path == required) {
            hopscotch.nextStep();
            return true;
          }
        }
      });
    } else if (category == "filter") {
      $.each(currentBlocks, function(k, v) {
        if (this.Type == "filter") {
          if (this.Rule.Filter == required) {
            hopscotch.nextStep();
            return true;
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
        return true;
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
      content: "<p>First, we need a <span class=\"tutorial-blockname\">fromhttpstream</span> block.</p><p>Click the hamburger button to see a list of every block in streamtools.</p>",
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
      content: "<p>Click <span class=\"tutorial-blockname\">fromhttpstream</span> to add that block, then click Next.</p>",
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
      content: "<p>You can click and drag blocks to move them around on screen.</p><p>Double-click the block to edit its parameters.</p><p>Type <span class=\"tutorial-url\">http://developer.usa.gov/1usagov</span> into the endpoint. Click Update, then click Next.</p>",
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
      content: "<p>Now let's add a block to log our data.</p><p>Double-click anywhere on screen to add a block.</p><p>Double-click anywhere on screen to add a block.</p><p>Type in <span class=\"tutorial-blockname\">tolog</span> and hit Enter.</p>",
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
      content: "<p>Let's connect the two, so we have data streaming into our log.</p><p>Click the OUT box on your <span class=\"tutorial-blockname\">fromhttpstream</span> box (the bottom black box).</p><p>Connect it to the IN on your <span class=\"tutorial-blockname\">tolog</span> (the top black box).</p>",
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

    ]
  };

// Start the tour!
hopscotch.startTour(tour, 0);


})();
