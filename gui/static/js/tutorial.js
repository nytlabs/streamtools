$(window).load(function() {
	var tour;

	tour = new Shepherd.Tour({
		defaults: {
			classes: 'shepherd-theme-arrows',
			scrollTo: true
		}
	});

	tour.addStep('welcome', {
		text: 'Welcome to Streamtools!',
		buttons: [
			{
				text: 'Next',
				action: tour.next
			}
		]
	});

	tour.addStep('goal', {
		text: 'In this demo, we\'ll use streamtools to see live clicks on the US government short links.',
		buttons: [
			{
				text: 'Next',
				action: tour.next
			}
		]
	});

	var clickRef = tour.addStep('intro-to-ref', {
		text: 'First, we need a fromhttpstream block.' + ' Click the hamburger button to see the reference.',
		attachTo: '#ui-ref-toggle',
		buttons: false
	});

	$("#ui-ref-toggle").one('click', function() {
		if (clickRef.isOpen()) {
			return Shepherd.activeTour.next();
		}
	});


	var addFromHTTP = tour.addStep('add-fromhttp', {
		text: 'Click fromhttpstream to add that block, then click Next.',
		attachTo: '#ui-ref-toggle',
		buttons: [
			{
				text: 'Next'
			}
		]
	});

	var editFromHTTP = tour.addStep('edit-fromhttp', {
		text: 'Double-click the block to edit its rules. Paste http://developer.usa.gov/1usagov into the endpoint, then click Next.',
		attachTo: '#ui-ref-toggle',
		buttons: [
			{
				text: 'Next'
			}
		]
	});

	var addTolog = tour.addStep('add-tolog', {
		text: 'Now let\'s add a block to log our data. Double-click anywhere on screen, then type in tolog.',
		attachTo: '#ui-ref-toggle',
		buttons: [
			{
				text: 'Next'
			}
		]
	});

	var makeConnection1 = tour.addStep('make-connection1', {
		text: 'Click the OUT box on your fromhttpstream box (the bottom black box), and connect it to the IN on your tolog (the top black box).',
		attachTo: '#ui-ref-toggle',
		buttons: [
			{
				text: 'Next'
			}
		]
	});

	var viewLog = tour.addStep('view-log', {
		text: 'Now click the log to view your data!',
		attachTo: '#log top',
		buttons: [
			{
				text: 'Next'
			}
		]
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
					return false;
				}
			});
		} else if (category == "endpoint") {
			$.each(currentBlocks, function(k, v) {
				console.log(this.Rule.Endpoint)
				if (this.Rule.Endpoint == required) {
					Shepherd.activeTour.next();
					return false;
				}
			});
		}
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
				return false;
	    	}
	    });
	}

	$(document).on("click", ".shepherd-button", function() {
		if (addFromHTTP.isOpen()) {
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
});