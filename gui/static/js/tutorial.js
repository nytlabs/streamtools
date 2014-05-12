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
    var tour, citibikeTour, govTour;

    // TODO: figure out a better way to do this so it doesn't flash on the screen
    // hide the intro text if it's on the page
    if ( $(".intro-text").length > 0) {
      $(".intro-text").remove();
    }

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
          console.log(this.Rule.Endpoint);
          if (this.Rule.Endpoint == required) {
            Shepherd.activeTour.next();
            return true;
          }
        });
      } else if (category == "interval") {
        $.each(currentBlocks, function(k, v) {
          console.log(this.Rule.Interval);
          if (this.Rule.Interval == required) {
            console.log("interval matches");
            Shepherd.activeTour.next();
            return true;
          }
        });
      } else if (category == "map") {
          $.each(currentBlocks, function(k, v) {
            if (this.Type == "map") {
              if (this.Rule.Map.url == required) {
                Shepherd.activeTour.next();
                return true;
              }
            }
        });
      } else if (category == "path") {
          $.each(currentBlocks, function(k, v) {
            if (this.Type == "gethttp" || this.Type == "unpack") {
              if (this.Rule.Path == required) {
                Shepherd.activeTour.next();
                return true;
              }
            }
        });
      } else if (category == "filter") {
          $.each(currentBlocks, function(k, v) {
            if (this.Type == "filter") {
              if (this.Rule.Filter == required) {
                Shepherd.activeTour.next();
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
          Shepherd.activeTour.next();
          return true;
        }
      });
    }

citibikeTour = new Shepherd.Tour({
  defaults: {
    classes: 'shepherd-theme-arrows',
    scrollTo: true
  }
});

citibikeTour.addStep('welcome', {
  text: 'Welcome to Streamtools.',
  tetherOptions: {
    targetAttachment: 'middle center',
  attachment: 'middle center',
  },
    attachTo: 'svg'
});

citibikeTour.addStep('goal', {
  text: 'In this demo, we\'ll use streamtools to get a live stream of Citibike availability--specifically, the station outside the NYT headquarters in Midtown Manhattan.',
    tetherOptions: {
      targetAttachment: 'middle center',
    attachment: 'middle center',
    },
    attachTo: 'svg'
});

citibikeTour.addStep('intro-to-ref', {
  text: [
    'We\'ll want to query our data source on regular intervals. We can use a <span class="tutorial-blockname">ticker</span> to do this.', 
    ' Click the hamburger button to see the list of every block in streamtools.'
    ],
    attachTo: '#ui-ref-toggle'
});

citibikeTour.addStep('add-ticker', {
  text: 'Click <span class="tutorial-blockname">ticker</span> to add that block, then click Next.',
  attachTo: '#ui-ref-toggle',
  buttons: [
    {
      text: "Next",
      action: checkBlockBeforeProgress("ticker", "type")
    }
  ]
});

citibikeTour.addStep('edit-ticker', {
  text: [
  'You can click and drag blocks to move them around on screen.',
  'Double-click the block to edit its parameters.', 
  'Let\'s set our interval to 10 seconds. Type <span class="tutorial-url">10s</span> into the Interval box and click Update.',
  'After that, click Next.'
  ],
    tetherOptions:
{
    targetAttachment: 'bottom left',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action: checkBlockBeforeProgress("10s", "interval")
    }
  ]
});

citibikeTour.addStep('add-map', {
  text: [
  'Before we can start making GET requests, we need to specify the URL from which we\'re getting the data.',
  'We\'ll use a <span class="tutorial-blockname">map</span> block for this, mapping the key "url" to our url.', 
  'Double-click anywhere on screen to add a block.',
  'Type in <span class="tutorial-blockname">map</span> and hit Enter.'
  ],
    tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action: checkBlockBeforeProgress("map", "type")
    }
  ]
});

citibikeTour.addStep('edit-map', {
  text: [
    'Double-click the map to edit its parameters.',
    'The <span class="tutorial-blockname">map</span> block takes <a href="https://github.com/nytlabs/gojee" target="_new">gojee</a> expression. Our map will look like this:',
    '<span class="tutorial-url">{</span>',
    '<span class="tutorial-url">\"url\": \"\'http://citibikenyc.com/stations/json\'\"</span>',
    '<span class="tutorial-url">}</span>',
    'Put that in the Map parameter, then click Next.'
  ],
  tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action: checkBlockBeforeProgress("\'http://citibikenyc.com/stations/json\'", "map")
    }
  ]
});

citibikeTour.addStep('make-connection1', {
  text: [
  'Let\'s connect the two, so every 10s, we map this URL.', 
  'Click the OUT box on your <span class="tutorial-blockname">ticker</span> box (the bottom black box). ' ,'Connect it to the IN on your <span class="tutorial-blockname">map</span> (the top black box).'
  ],
    tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action: checkConnectionsBeforeProgress("ticker", "map")
    }
  ]
});

citibikeTour.addStep('add-HTTP', {
  text: [
  'Now we need to actually get our data. We\'ll make this GET request with a <span class="tutorial-blockname">gethttp</span> block.', 
  'Double-click anywhere on screen to add a block.',
  'Type in <span class="tutorial-blockname">gethttp</span> and hit Enter.'
  ],
    tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action: checkBlockBeforeProgress("gethttp", "type")
    }
  ]
});

citibikeTour.addStep('edit-http', {
  text: [
    'Double-click on your <span class="tutorial-blockname">gethttp</span> block to edit it.',
    'Our URL is mapped to the path <span class="tutorial-url">.url</span>.',
    'Put that in the Path parameter, then click Next.'
  ],
  tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action:  checkBlockBeforeProgress(".url", "path")
    }
  ]
});

citibikeTour.addStep('make-connection2', {
  text: [
  'Now let\s connect our <span class="tutorial-blockname">map</span> block to our <span class="tutorial-blockname">gethttp</span> block.', 
  'That way, we\'ll make a GET request to that URL every 10s.',
  'Click the OUT box on your <span class="tutorial-blockname">ticker</span> box (the bottom black box). ',
  'Connect it to the IN on your <span class="tutorial-blockname">gethttp</span> (the top black box).'
  ],
    tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action:  checkConnectionsBeforeProgress("map", "gethttp")
    }
  ]
});

citibikeTour.addStep('add-unpack', {
  text: [
  'If you view the <a href="http://citibikenyc.com/stations/json" target="_new">JSON data</a> in your browser, you\'ll see that all the data is in a big array.', 
  'The key wrapping up all the data about individual stations is <span class="tutorial-url">stationBeanList</span>.',
  'In order to be able to manipulate and filter this data, we need to unpack it first.',
  'That\'s where the <span class="tutorial-blockname">unpack</span> block comes in handy. Double-click and create it anywhere on-screen.'
  ],
    tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action:  checkBlockBeforeProgress("unpack", "type")
    }
  ]
});

citibikeTour.addStep('edit-unpack', {
  text: [
  'Double-click on the <span class="tutorial-blockname">unpack</span> block to edit its rule.', 
  'Set its Path to <span class="tutorial-url">.stationBeanList</span> and click Next.',
  ],
    tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action:  checkBlockBeforeProgress(".stationBeanList", "path")
    }
  ]
});

citibikeTour.addStep('make-connection3', {
  text: [
  'Now let\'s connect our <span class="tutorial-blockname">gethttp</span> (the thing giving us the JSON) to our <span class="tutorial-blockname">unpack</span> (the thing iterating over that JSON).', 
  'Connect the two and click Next.',
  ],
    tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action:  checkConnectionsBeforeProgress("gethttp", "unpack")
  
    }
  ]
});

citibikeTour.addStep('add-filter', {
  text: [
  'Right now we\'re getting data about every station. Let\'s filter out every station other than the one outside the NYT headquarters.', 
  'For this, we\'ll use a <span class="tutorial-blockname">filter</span> block.',
  'Click Next once you\'ve made it.',
  ],
    tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action:  checkBlockBeforeProgress("filter", "type")
    }
  ]
});

citibikeTour.addStep('edit-filter', {
  text: [
  'The station nearest the NYT HQ is <span class="tutorial-url">\'W 41st St & 8 Ave\'</span>.', 
  'Our <span class="tutorial-blockname">filter</span> rule will look like this:',
  '<span class="tutorial-url">.stationName == \'W 41 St & 8 Ave\'</span>',
  ],
    tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action: checkBlockBeforeProgress(".stationName == 'W 41 St & 8 Ave'", "filter")
  
    }
  ]
});

citibikeTour.addStep('make-connection4', {
  text: [
  'Connect your <span class="tutorial-blockname">unpack</span> and <span class="tutorial-blockname">filter</span> blocks.', 
  ],
    tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action: checkConnectionsBeforeProgress("unpack", "filter")
    }
  ]
});

citibikeTour.addStep('add-tolog', {
  text: [
  'A quick and easy way to see your data stream is to log it using a <span class="tutorial-blockname">tolog</span> block.', 
  'The <span class="tutorial-blockname">tolog</span> block logs your data to the console and the log built into streamtools.',
  'Add it and click Next.',
  ],
    tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action: checkBlockBeforeProgress("tolog", "type")
    }
  ]
});

citibikeTour.addStep('make-connection5', {
  text: [
  'Finally, connect your <span class="tutorial-blockname">filter</span> and <span class="tutorial-blockname">tolog</span> blocks.', 
  ],
    tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
  },
    attachTo: 'svg',
  buttons: [
    {
      text: "Next",
      action: checkConnectionsBeforeProgress("filter", "tolog")
    }
  ]
});

citibikeTour.addStep('finished', {
  text: [
  'Now, every 10s, your log will be updated with your newest filtered live data.', 
  ],
    tetherOptions:
{
  targetAttachment: 'bottom right',
    attachment: 'bottom right',
},
buttons: [ { text: 'Complete', action: Shepherd.activeTour.complete() } ],
    attachTo: 'svg'
});

govTour = new Shepherd.Tour({
  defaults: {
    classes: 'shepherd-theme-arrows',
    scrollTo: true
  }
});

govTour.addStep("welcome", {
  text: "Welcome to Streamtools!",
  tetherOptions: { "targetAttachment": "middle center", "attachment": "middle center" },
    attachTo: 'svg'
});
 govTour.addStep("goal", { 
   text: 'In this demo, we\'ll use streamtools to see live clicks on the US government short links.',
   tetherOptions: { "targetAttachment": "middle center", "attachment": "middle center" },
     attachTo: 'svg'
 });
 govTour.addStep("intro-to-ref", {
   text: ['First, we need a <span class="tutorial-blockname">fromhttpstream</span> block.' , ' Click the hamburger button to see the reference.'],
   attachTo: '#ui-ref-toggle'
 });
 govTour.addStep("add-fromhttp", { 
   text: 'Click <span class="tutorial-blockname">fromhttpstream</span> to add that block, then click Next.',
   attachTo: 'li[data-block-type="fromhttpstream"]'
 });
 govTour.addStep("edit-fromhttp", { 
   text: ['Double-click the block to edit its rules.', 'Paste <span class="tutorial-url">http://developer.usa.gov/1usagov</span> into the endpoint, then click Next.'],
     attachTo: 'svg'
 });
 govTour.addStep("add-tolog", { 
   text: [ 'Now let\'s add a block to log our data.', 'Double-click anywhere on screen to add a block.', 'Type in <span class="tutorial-blockname">tolog</span> and hit Enter.' ],
   tetherOptions: { 'targetAttachment': 'bottom right', 'attachment': 'bottom right' },
     attachTo: 'svg'
 });
 govTour.addStep("make-connection1", { 
   text: [ 'Let\'s connect the two, so we have data streaming into our log.', 'Click the OUT box on your <span class="tutorial-blockname">fromhttpstream</span> box (the bottom black box). ' ,'Connect it to the IN on your <span class="tutorial-blockname">tolog</span> (the top black box).' ],
   tetherOptions: { 'targetAttachment': 'bottom right', 'attachment': 'bottom right' },
     attachTo: 'svg'
 });
 govTour.addStep("view-log", { 
   text: 'Now click the log (this black bar) to view your data!',
   tetherOptions: { 'targetAttachment': 'bottom center', 'attachment': 'bottom center' },
   buttons: [{ 'text': 'Complete' } ],
     attachTo: 'svg'
 });

    tours = { 
      'citibike': citibikeTour,
      'gov': govTour
    };

    tour = tours[params['tutorial']];
    if (tour) {
      tour.start();
    }

// move these functions to buttons [{}] on each step
//$(document).on("click", ".shepherd-button", function() {
//  else if ( govTour.getById('welcome').isOpen() || govTour.getById('goal').isOpen() || govTour.getById('intro-to-ref').isOpen() ) {
//    Shepherd.activeTour.next();
//  }
//  else if ( govTour.getById('add-fromhttp').isOpen() ) {
//    var b = $("text:contains('fromhttpstream')").prev();
//    httpBlock = "rect[data-id='" + b.attr('data-id') + "']";
//
//    govTour.getById("edit-fromhttp")["options"]["attachTo"] = httpBlock;
//    checkBlockBeforeProgress("fromhttpstream", "type");
//
//  } 
//  else if ( govTour.getById('edit-fromhttp').isOpen() ) {
//    checkBlockBeforeProgress("http://developer.usa.gov/1usagov", "endpoint");
//  } 
//  else if ( govTour.getById('add-tolog').isOpen() ) {
//    checkBlockBeforeProgress("tolog", "type");
//  }
//  else if (govTour.getById('make-connection1').isOpen()) {
//    checkConnectionsBeforeProgress("fromhttpstream", "tolog");
//  }
//  else if (govTour.getById('view-log').isOpen()) {
//    Shepherd.activeTour.complete();	
//  }
//});
    
  }

});
