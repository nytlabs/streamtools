// http://www.paulirish.com/2009/throttled-smartresize-jquery-event-handler/
(function($, sr) {

    // debouncing function from John Hann
    // http://unscriptable.com/index.php/2009/03/20/debouncing-javascript-methods/
    var debounce = function(func, threshold, execAsap) {
        var timeout;

        return function debounced() {
            var obj = this,
                args = arguments;

            function delayed() {
                if (!execAsap)
                    func.apply(obj, args);
                timeout = null;
            }

            if (timeout)
                clearTimeout(timeout);
            else if (execAsap)
                func.apply(obj, args);

            timeout = setTimeout(delayed, threshold || 100);
        };
    };
    // smartresize 
    jQuery.fn[sr] = function(fn) {
        return fn ? this.bind('resize', debounce(fn)) : this.trigger(sr);
    };
})(jQuery, 'smartresize');

$(function() {

    var library = d3.nest()
        .key(function(d, i){return d.Type})
        .rollup(function(d){return d[0]})
        .map(JSON.parse($.ajax({
            url: '/library',
            type: 'GET',
            async: false
        }).responseText).Library)

    console.log(library)

    var blocks = [];
    var connections = [];

    var width = $(window).width(),
        height = $(window).height();

    var svg = d3.select("body").append("svg")
        .attr("width", width)
        .attr("height", height);

    var linkContainer = svg.append('g')
        .attr('class', 'linkContainer');

    var nodeContainer = svg.append('g')
        .attr('class', 'nodeContainer');

    var link = svg.select(".linkContainer").selectAll(".link"),
        node = svg.select(".nodeContainer").selectAll(".node");

    var drag = d3.behavior.drag()
        .on("drag", function(d,i) {
            d.Position.X += d3.event.dx
            d.Position.Y += d3.event.dy
            d3.select(this).attr("transform", function(d,i){
                return "translate(" + [ d.Position.X, d.Position.Y ] + ")"
            })
            updateLinks();
        })
        .on("dragend", function(d, i){
            $.ajax({
                url: '/blocks/' + d.Id,
                type: 'PUT',
                data: JSON.stringify(d.Position),
                success: function(result) {}
            });
        })

    function logReader() {
        var logTemplate = $('#log-item-template').html();
        this.ws = new WebSocket("ws://localhost:7070/log");

        this.ws.onmessage = function(d) {
            var logData = JSON.parse(d.data);
            var logItem = $("<div />").addClass("log-item");
            $("#log").append(logItem);

            var tmpl = _.template(logTemplate, {
                item: {
                    type: logData.Type,
                    time: new Date(),
                    data: logData.Data,
                    id: logData.Id,
                }
            });

            logItem.html(tmpl);

            var log = document.getElementById('log');
            log.scrollTop = log.scrollHeight;
        };
    }

    function uiReader() {
        _this = this;
        _this.handleMsg = null;
        this.ws = new WebSocket("ws://localhost:7070/ui");
        this.ws.onopen = function(d) {
            _this.ws.send("get_state");
        };
        this.ws.onmessage = function(d) {
            var uiMsg = JSON.parse(d.data)
            var isBlock = uiMsg.Data.hasOwnProperty('Type');
            switch (uiMsg.Type) {
                case "CREATE":
                    if(isBlock){
                        uiMsg.Data.TypeInfo = library[uiMsg.Data.Type]
                        blocks.push(uiMsg.Data)
                    } else {
                        connections.push(uiMsg.Data)
                    }
                    update();
                    break
                case "DELETE":
                    if(isBlock){
                        for(var i = 0; i < blocks.length; i++){
                            blocks.splice(i, 1)
                        }
                    } else {
                        for(var i = 0; i < connections.length; i++){
                            connections.splice(i, 1)
                        }
                    }
                    update();
                    break
                case "UPDATE":
                    if(isBlock){
                        var block = null;
                        for(var i = 0; i < blocks.length; i++){
                            if(blocks[i].Id === uiMsg.Data.Id){
                                block = blocks[i];
                                break;
                            }
                        }
                        if(block !== null){
                            block.Position = uiMsg.Data.Position
                            update();
                        }

                    }
                    updateLinks();
                    break
                case "QUERY":
                    break
            }
        };
    }

    function update(){
        node = node.data(blocks, function(d) {
            return d.Id;
        });

        var nodes = node.enter()
            .append("g")
            .call(drag)
   
        var rects = nodes.append("rect")
            .attr("class", "node")

        var idRects = nodes.append("rect")
            .attr('class', 'idrect')

        nodes.append("svg:text")
            .attr("class", "nodetype")
            .attr("dx", 0)
            .text(function(d) {
                return d.Type;
            }).each(function(d) {
                var bbox = this.getBBox();
                d.width = (d.width > bbox.width ? d.width : bbox.width + 30);
                d.height = (d.height > bbox.height ? d.height : bbox.height+ 5);
            }).attr("dy", function(d) {
                return 1 * d.height + 5;
            })

        /*nodes.append("svg:text")
            .attr("class", "nodeid")
            .attr("dx", 0)
            .attr("dy", function(d) {
                return 1 * d.height;
            })
            .text(function(d) {
                return d.Id;
            }).each(function(d) {
                var bbox = this.getBBox();
                d.width = (d.width > bbox.width + 20 ? d.width : bbox.width + 20);
                d.height = (d.height > bbox.height + 20 ? d.height : bbox.height + 20);
            });*/

        idRects
            .attr('x', 0)
            .attr('y', 0)
            .attr('width', function(d) {
                return d.width;
            })
            .attr('height', function(d) {
                return d.height * 2;
            })

        /*rects
            .attr('x', 0)
            .attr('y', 0)
            .attr('width', function(d) {
                return d.width;
            })
            .attr('height', function(d) {
                return d.height * 2;
            })*/

        node.attr("transform", function(d) {
            return "translate(" + d.Position.X + ", " + d.Position.Y + ")";
        });

        var inRoutes = node.selectAll('.inRoutes')
            .data(function(d){
                return d.TypeInfo.InRoutes
            })

        inRoutes.enter()
            .append("rect")
            .attr("class","chan-in")
            .attr("x",function(d, i){
                return i * 15
            })
            .attr("y",0)
            .attr("width", 10)
            .attr("height",10)

        inRoutes.exit().remove();

        var queryRoutes = node.selectAll('.queryRoutes')
            .data(function(d){
                return d.TypeInfo.QueryRoutes
            })

        queryRoutes.enter()
            .append("rect")
            .attr("class","chan-in")
            .attr("x",function(d, i){
                var p = d3.select(this.parentNode).datum()
                return (p.width - 10)
            })
            .attr("y",function(d, i){
                return i * 15
            })
            .attr("width", 10)
            .attr("height",10)

        queryRoutes.exit().remove();

        var outRoutes = node.selectAll('.outRoutes')
            .data(function(d){
                return d.TypeInfo.OutRoutes
            })

        outRoutes.enter()
            .append("rect")
            .attr("class","chan-in")
            .attr("x",function(d, i){
                return i * 15
            })
            .attr("y",function(d, i){
                var p = d3.select(this.parentNode).datum()
                return ((p.height * 2) - 10)
            })
            .attr("width", 10)
            .attr("height",10)

        outRoutes.exit().remove();

        node.exit().remove();

        link = link.data(connections, function(d){
            return d.Id
        })

        link.enter()
            .append("line")
            .attr("class", "link")

        updateLinks();

        link.exit().remove();
    }

    // updateLinks() is too slow!
    function updateLinks(){
        link.attr("x1", function(d){
                var from = node.filter(function(p, i){ return p.Id == d.FromId }).datum()
                return from.Position.X + 5
            })
            .attr("y1", function(d){
                var from = node.filter(function(p, i){ return p.Id == d.FromId }).datum()
                return (from.Position.Y + from.height * 2) - 5
            })
            .attr("x2", function(d){
                var to = node.filter(function(p, i){ return p.Id == d.ToId }).datum()
                var i = to.TypeInfo.InRoutes.indexOf(d.ToRoute)
                return to.Position.X + (i * 15) + 5 
            })
            .attr("y2", function(d){
                var to = node.filter(function(p, i){ return p.Id == d.ToId }).datum()
                return to.Position.Y + 5
            })
    }

    b = new logReader();
    c = new uiReader();







});