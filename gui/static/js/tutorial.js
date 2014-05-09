$(window).load(function() {
  // from http://stackoverflow.com/a/647272
  // until we figure out a better/different way to trigger tutorials
  // i'm gonna go with query string params (linkable, easy to do)
  // this function parses the url after ? into key/value pairs
  function queryParams() {
    var result = {}, keyValuePairs = location.search.slice(1).split('&');

    keyValuePairs.forEach(function(keyValuePair) {
      keyValuePair = keyValuePair.split('=');
      result[keyValuePair[0]] = keyValuePair[1] || '';
    });

    return result;
  }
  
  // grab the query string params
  params = queryParams();

  // does the url have params that include 'tutorial'? if so, load up the... tutorial.
  // otherwise just skip all this, streamtools as usual.
  if (params && params["tutorial"]) {

    // TODO: figure out a better way to do this so it doesn't flash on the screen
    // hide the intro text if it's on the page
    if ( $(".intro-text").length > 0) {
      $(".intro-text").remove();
    }

    var tour;
    var httpBlock;

    tour = new Shepherd.Tour({
      defaults: {
        classes: 'shepherd-theme-arrows',
         scrollTo: true
      }
    });

    var welcome = tour.addStep('welcome', {
      text: 'Welcome to Streamtools!',
        attachTo: 'svg',
        tetherOptions: {
          targetAttachment: 'middle center',
        attachment: 'middle center',
        },
        buttons: [
    {
      text: 'Next',
    }
    ]
    });

    var goal = tour.addStep('goal', {
      text: 'In this demo, we\'ll use streamtools to see live clicks on the US government short links.',
        attachTo: 'svg',
        tetherOptions: {
          targetAttachment: 'middle center',
        attachment: 'middle center',
        },
        buttons: [
    {
      text: 'Next',
    }
    ]
    });

    var clickRef = tour.addStep('intro-to-ref', {
      text: ['First, we need a <span class="tutorial-blockname">fromhttpstream</span> block.' , ' Click the hamburger button to see the reference.'],
        attachTo: '#ui-ref-toggle',
        buttons: false
    });

    var addFromHTTP = tour.addStep('add-fromhttp', {
      text: 'Click <span class="tutorial-blockname">fromhttpstream</span> to add that block, then click Next.',
        attachTo: 'li[data-block-type="fromhttpstream"]',
        buttons: [
    {
      text: 'Next'
    }
    ],
    });

    $("#ui-ref-toggle").one('click', function() {
      if (clickRef.isOpen()) {
        return Shepherd.activeTour.next();
      }
    });

    var editFromHTTP = tour.addStep('edit-fromhttp', {
      text: [
      'Double-click the block to edit its rules.', 
        'Paste <span class="tutorial-url">http://developer.usa.gov/1usagov</span> into the endpoint, then click Next.'
      ],
        tetherOptions:
    {
      targetAttachment: 'top left',
        attachment: 'top right',
    },
        attachTo: httpBlock,
        buttons: [
    {
      text: 'Next'
    }
    ],
    });

    var addTolog = tour.addStep('add-tolog', {
      text: [
      'Now let\'s add a block to log our data.', 
        'Double-click anywhere on screen to add a block.',
        'Type in <span class="tutorial-blockname">tolog</span> and hit Enter.'
      ],
        tetherOptions:
    {
      targetAttachment: 'bottom right',
        attachment: 'bottom right',
    },
        attachTo: 'svg',
        buttons: [
    {
      text: 'Next'
    }
    ]
    });

    var makeConnection1 = tour.addStep('make-connection1', {
      text: [
      'Let\'s connect the two, so we have data streaming into our log.', 
        'Click the OUT box on your <span class="tutorial-blockname">fromhttpstream</span> box (the bottom black box). ' ,'Connect it to the IN on your <span class="tutorial-blockname">tolog</span> (the top black box).'
      ],
        tetherOptions:
    {
      targetAttachment: 'bottom right',
        attachment: 'bottom right',
    },
        attachTo: 'svg',
        buttons: [
    {
      text: 'Next'
    }
    ]
    });

    var viewLog = tour.addStep('view-log', {
      text: 'Now click the log (this black bar) to view your data!',
        tetherOptions:
    {
      targetAttachment: 'bottom center',
        attachment: 'bottom center',
    },
        attachTo: "svg",
        buttons: [
    {
      text: 'Complete'
    }
    ],
    });

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
            Shepherd.activeTour.next();
            return true;
          }
        });
      } else if (category == "endpoint") {
        $.each(currentBlocks, function(k, v) {
          console.log(this.Rule.Endpoint)
          if (this.Rule.Endpoint == required) {
            Shepherd.activeTour.next();
            return true;
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
          Shepherd.activeTour.next();
          return true;
        }
      });
    }

    $(document).on("click", ".shepherd-button", function() {
      if (welcome.isOpen()) {
        Shepherd.activeTour.next();
      }
      else if (goal.isOpen()) {
        Shepherd.activeTour.next();
      }
      else if (addFromHTTP.isOpen()) {
        var b = $("text").text("fromhttpstream").prev();
        httpBlock = "rect[data-id='" + b.attr('data-id') + "']";

        tour.getById("edit-fromhttp")["options"]["attachTo"] = httpBlock;
        checkBlockBeforeProgress("fromhttpstream", "type");
      } 
      else if (editFromHTTP.isOpen()) {
        checkBlockBeforeProgress("http://developer.usa.gov/1usagov", "endpoint");
      } 
      else if (addTolog.isOpen()) {
        checkBlockBeforeProgress("tolog", "type");
      }
      else if (makeConnection1.isOpen()) {
        checkConnectionsBeforeProgress("fromhttpstream", "tolog");
      }
      else if (viewLog.isOpen()) {
        Shepherd.activeTour.complete();	
      }
    });
    tour.start();
  }

});
