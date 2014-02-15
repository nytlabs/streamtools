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
        .key(function(d, i) {
            return d.Type;
        })
        .rollup(function(d) {
            return d[0];
        })
        .map(JSON.parse($.ajax({
            url: '/library',
            type: 'GET',
            async: false
        }).responseText).Library);

    var blocks = [];
    var connections = [];

    var width = window.innerWidth,
        height = window.innerHeight;

    var mouse = {
        x: 0,
        y: 0
    };

    var multiSelect = false;

    $(window).mousemove(function(e){
        mouse = {
            x: e.clientX,
            y: e.clientY
        };
        if (isConnecting) {
            updateNewConnection();
        } 
    })

    $(window).keydown(function(e){
        // check to see if any text box is selected
        // if so, don't allow multiselect
        if( $('input').is(':focus') ) {
            return;
        }

        // if key is backspace or delete
        if (e.keyCode == 8 || e.keyCode == 46){
            e.preventDefault();
            d3.selectAll('.selected')
                .each(function(d){
                    if(this.classList.contains('idrect')){
                        $.ajax({
                            url: '/blocks/' + d3.select(this.parentNode).datum().Id,
                            type: 'DELETE',
                            success: function(result) {}
                        });
                    }
                    if(this.classList.contains('rateLabel')){
                        $.ajax({
                            url: '/connections/' + d3.select(this).datum().Id,
                            type: 'DELETE',
                            success: function(result) {}
                        });
                    }
                })
        }

        multiSelect = e.shiftKey
    })

    $(window).keyup(function(e){
        multiSelect = e.shiftKey
    })

    var svg = d3.select("body").append("svg")
        .attr("width", width)
        .attr("height", height)

    var bg = svg.append('rect')
        .attr('x', 0)
        .attr('y', 0)
        .attr('class', 'background')
        .attr('width', width)
        .attr('height', height)
        .on("dblclick", function() {
            $("#create")
                .css({
                    top: mouse.y,
                    left: mouse.x,
                    "visibility": "visible"
                });
            $("#create-input").focus();
        })
        .on("click", function() {
            if (isConnecting) {
                terminateConnection();
            }
        })
        .on("mousedown", function(){
            d3.selectAll('.selected')
                .classed('selected', false) 
        })

    $(window).smartresize(function(e) {
        svg.attr("width", window.innerWidth)
            .attr("height", window.innerHeight);
        bg.attr('width', window.innerWidth)
            .attr('height', window.innerHeight)
    });

    var linkContainer = svg.append('g')
        .attr('class', 'linkContainer');

    var nodeContainer = svg.append('g')
        .attr('class', 'nodeContainer');

    var link = svg.select(".linkContainer").selectAll(".link"),
        node = svg.select(".nodeContainer").selectAll(".node");

    var tooltip = d3.select("body")
        .append("div")
        .attr('class', 'tooltip');

    var drag = d3.behavior.drag()
        .on("drag", function(d, i) {
            d.Position.X += d3.event.dx;
            d.Position.Y += d3.event.dy;
            d3.select(this)
                .attr("transform", function(d, i) {
                    return "translate(" + [d.Position.X, d.Position.Y] + ")";
                })
            updateLinks();
        })
        .on("dragend", function(d, i) {
            $.ajax({
                url: '/blocks/' + d.Id,
                type: 'PUT',
                data: JSON.stringify(d.Position),
                success: function(result) {}
            });
        });

    var newConnection = svg.select('.linkcontainer').append('path')
        .attr("id", "newLink")
        .style("fill", "none")
        .on("click", function() {
            if (isConnecting) {
                terminateConnection();
            }
        });

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
            var uiMsg = JSON.parse(d.data);
            var isBlock = uiMsg.Data.hasOwnProperty('Type');
            switch (uiMsg.Type) {
                case "CREATE":
                    if (isBlock) {
                        uiMsg.Data.TypeInfo = library[uiMsg.Data.Type];
                        blocks.push(uiMsg.Data);
                    } else {
                        connections.push(uiMsg.Data);
                    }
                    update();
                    break;
                case "DELETE":
                    for(var i = 0; i < blocks.length; i++){
                        if(uiMsg.Data.Id == blocks[i].Id){
                            blocks.splice(i, 1)
                            break;
                        }
                    }
                    for(var i = 0; i < connections.length; i++){
                        if(uiMsg.Data.Id == connections[i].Id){
                            connections.splice(i, 1)
                            break;
                        }
                    }
                    update();
                    break;
                case "UPDATE":
                    if (isBlock) {
                        var block = null;
                        for (var i = 0; i < blocks.length; i++) {
                            if (blocks[i].Id === uiMsg.Data.Id) {
                                block = blocks[i];
                                break;
                            }
                        }
                        if (block !== null) {
                            block.Position = uiMsg.Data.Position;
                            update();
                        }

                    }
                    updateLinks();
                    break;
                case "QUERY":
                    break;
            }
        };
    }

    var d3line2 = d3.svg.line()
        .x(function(d) {
            return d.x;
        })
        .y(function(d) {
            return d.y;
        })
        .interpolate("monotone");

    function update() {
        node = node.data(blocks, function(d) {
            return d.Id;
        });

        var nodes = node.enter()
            .append("g")
            .attr("class", "node")
            .call(drag);

        //var rects = nodes.append("rect")
        //    .attr("class", "node");

        var idRects = nodes.append("rect")
            .attr('class', 'idrect');

        nodes.append("svg:text")
            .attr("class", "nodetype unselectable")
            .attr("dx", 0)
            .text(function(d) {
                return d.Type;
            }).each(function(d) {
                var bbox = this.getBBox();
                d.width = (d.width > bbox.width ? d.width : bbox.width + 30);
                d.height = (d.height > bbox.height ? d.height : bbox.height + 5);
            }).attr("dy", function(d) {
                return 1 * d.height + 5;
            })
            .on('mousedown', function() {
                if (!multiSelect){
                    d3.selectAll('.selected')
                        .classed('selected', false)
                }
                d3.select(this.parentNode).select('.idrect')
                    .classed('selected', true) 
            });

        idRects
            .attr('x', 0)
            .attr('y', 0)
            .attr('width', function(d) {
                return d.width;
            })
            .attr('height', function(d) {
                return d.height * 2;
            })
            .on('mousedown', function() {
                if (!multiSelect){
                    d3.selectAll('.selected')
                        .classed('selected', false) 
                }
                d3.select(this)
                    .classed('selected', true) 
            });

        node.attr("transform", function(d) {
            return "translate(" + d.Position.X + ", " + d.Position.Y + ")";
        });

        var inRoutes = node.selectAll('.in')
            .data(function(d) {
                return d.TypeInfo.InRoutes;
            });

        inRoutes.enter()
            .append("rect")
            .attr("class", "chan in")
            .attr("x", function(d, i) {
                return i * 15;
            })
            .attr("y", 0)
            .attr("width", 10)
            .attr("height", 10)
            .on("mouseover", function(d) {
                tooltip.text(d);
                return tooltip.style("visibility", "visible");
            })
            .on("mousemove", function(d) {
                return tooltip.style("top", (event.pageY - 10) + "px").style("left", (event.pageX + 10) + "px");
            })
            .on("mouseout", function(d) {
                return tooltip.style("visibility", "hidden");
            }).on("click", function(d) {
                handleConnection(d3.select(this.parentNode).datum(), d, "in");
            });

        inRoutes.exit().remove();

        var queryRoutes = node.selectAll('.query')
            .data(function(d) {
                return d.TypeInfo.QueryRoutes;
            });

        queryRoutes.enter()
            .append("rect")
            .attr("class", "chan query")
            .attr("x", function(d, i) {
                var p = d3.select(this.parentNode).datum();
                return (p.width - 10);
            })
            .attr("y", function(d, i) {
                return i * 15;
            })
            .attr("width", 10)
            .attr("height", 10)
            .on("mouseover", function(d) {
                tooltip.text(d);
                return tooltip.style("visibility", "visible");
            })
            .on("mousemove", function(d) {
                return tooltip.style("top", (event.pageY - 10) + "px").style("left", (event.pageX + 10) + "px");
            })
            .on("mouseout", function(d) {
                return tooltip.style("visibility", "hidden");
            })

        queryRoutes.exit().remove();

        var outRoutes = node.selectAll('.out')
            .data(function(d) {
                return d.TypeInfo.OutRoutes;
            });

        outRoutes.enter()
            .append("rect")
            .attr("class", "chan out")
            .attr("x", function(d, i) {
                return i * 15;
            })
            .attr("y", function(d, i) {
                var p = d3.select(this.parentNode).datum();
                return ((p.height * 2) - 10);
            })
            .attr("width", 10)
            .attr("height", 10)
            .on("mouseover", function(d) {
                tooltip.text(d);
                return tooltip.style("visibility", "visible");
            })
            .on("mousemove", function(d) {
                return tooltip.style("top", (event.pageY - 10) + "px").style("left", (event.pageX + 10) + "px");
            })
            .on("mouseout", function(d) {
                return tooltip.style("visibility", "hidden");
            }).on("click", function(d) {
                handleConnection(d3.select(this.parentNode).datum(), d, "out");
            });

        outRoutes.exit().remove();

        node.exit().remove();

        link = link.data(connections, function(d) {
            return d.Id;
        });

        link.enter()
            .append("svg:path")
            .attr("class", "link")
            .style("fill", "none")
            .attr("id", function(d) {
                return "link_" + d.Id;
            })
            .each(function(d) {
                d.path = d3.select(this)[0][0];
                d.from = node.filter(function(p, i) {
                    return p.Id == d.FromId;
                }).datum();
                d.to = node.filter(function(p, i) {
                    return p.Id == d.ToId;
                }).datum();
                d.rate = 10.00;
                d.rateLoc = 0.0;
            });

        var ping = svg.select('.linkContainer').selectAll(".edgePing")
            .data(connections, function(d) {
                return d.Id;
            });

        ping.enter()
            .append("circle")
            .attr("class", "edgePing")
            .attr("r", 4);

        ping.exit().remove();

        var edgeLabel = svg.select('.linkContainer').selectAll(".edgeLabel")
            .data(connections, function(d) {
                return d.Id;
            });

        var ed = edgeLabel.enter()
            .append("g")
            .attr("class", "edgeLabel")
            .append("text")
            .attr("dy", -2)
            .attr("text-anchor", "middle")
            .append("textPath")
            .attr("class", "rateLabel unselectable")
            .attr("startOffset", "50%")
            .attr("xlink:href", function(d) {
                return "#link_" + d.Id;
            })
            .text(function(d) {
                return d.rate;
            })
            .on('mousedown', function() {
                if (!multiSelect){
                    d3.selectAll('.selected')
                        .classed('selected', false) 
                }
                d3.select(this)
                    .classed('selected', true) 
            });

        edgeLabel.exit().remove();

        updateLinks();
        link.exit().remove();
    }

    window.setInterval(function() {
        d3.selectAll(".rateLabel")
            .text(function(d) {
                // this is dumb.
                // d.rate = Math.sin(+new Date() * .0000001) * Math.random() * 5;
                return Math.round(100 * d.rate) / 100.0;
            });
    }, 100);

    window.setInterval(function() {
        svg.selectAll('.edgePing')
            .each(function(d) {
                d.rate += Math.random();
                d.rateLoc += 0.001 + Math.min(d.rate, 100) / 4000.0;
                if (d.rateLoc > 1) d.rateLoc = 0;
                d.edgePos = d.path.getPointAtLength(d.rateLoc * d.path.getTotalLength());
            })
            .attr('cx', function(d) {
                return d.edgePos.x;
            })
            .attr('cy', function(d) {
                return d.edgePos.y;
            });
    }, 1000 / 60);

    // updateLinks() is too slow!
    function updateLinks() {
        link.attr("d", function(d) {
            return d3line2([{
                x: d.from.Position.X + 5,
                y: (d.from.Position.Y + d.from.height * 2) - 5
            }, {
                x: d.from.Position.X + 5,
                y: (d.from.Position.Y + d.from.height * 2) + 15
            }, {
                x: d.to.Position.X + (d.to.TypeInfo.InRoutes.indexOf(d.ToRoute) * 15) + 5,
                y: d.to.Position.Y - 15
            }, {
                x: d.to.Position.X + (d.to.TypeInfo.InRoutes.indexOf(d.ToRoute) * 15) + 5,
                y: d.to.Position.Y + 5
            }]);
        });
    }

    $("#create-input").focusout(function() {
        $("#create-input").val('');
        $("#create").css({
            "visibility": "hidden"
        });
    });

    $("#create-input").keyup(function(k) {
        if (k.keyCode == 13) {
            createBlock();
            $("#create").css({
                "visibility": "hidden"
            });
            $("#create-input").val('');
        }
    });

    function createBlock() {
        var blockType = $("#create-input").val()
        if (!library.hasOwnProperty(blockType)) {
            return;
        }
        var offset = $("#create").offset()

        $.ajax({
            url: '/blocks',
            type: 'POST',
            data: JSON.stringify({
                "Type": blockType,
                "Position": {
                    "X": offset.left,
                    "Y": offset.top
                }
            }),
            success: function(result) {}
        });
    }

    var isConnecting = false;
    var newConn = {};

    function handleConnection(block, route, routeType) {
        isConnecting = !isConnecting;
        isConnecting ? startConnection(block, route, routeType) : endConnection(block, route, routeType);
    }

    function startConnection(block, route, routeType) {
        newConn = {
            start: block,
            startRoute: route,
            startType: routeType
        };
        updateNewConnection();
        newConnection.style("visibility", "visible");
    }

    function endConnection(block, route, routeType) {
        if (newConn.startType === routeType) {
            return;
        }

        var connReq = {
            "FromId": null,
            "ToId": null,
            "ToRoute": null
        }

        if (newConn.startType == "out") {
            connReq.FromId = newConn.start.Id;
            connReq.ToId = block.Id;
            connReq.ToRoute = route;
        } else {
            connReq.FromId = block.Id;
            connReq.ToId = newConn.start.Id;
            connReq.ToRoute = newConn.startRoute;
        }

        $.ajax({
            url: '/connections',
            type: 'POST',
            data: JSON.stringify(connReq),
            success: function(result) {}
        });

        terminateConnection();
    }

    function updateNewConnection() {
        newConnection.attr('d', function() {
            return d3line2(newConn.startType == "out" ?
                [{
                    x: newConn.start.Position.X + 5,
                    y: (newConn.start.Position.Y + newConn.start.height * 2) - 5
                }, {
                    x: newConn.start.Position.X + 5,
                    y: (newConn.start.Position.Y + newConn.start.height * 2) + 15
                }, {
                    x: mouse.x,
                    y: mouse.y - 15
                }, {
                    x: mouse.x,
                    y: mouse.y
                }] :
                [{
                    x: newConn.start.Position.X + (newConn.start.TypeInfo.InRoutes.indexOf(newConn.startRoute) * 15) + 5,
                    y: newConn.start.Position.Y + 5
                }, {
                    x: newConn.start.Position.X + (newConn.start.TypeInfo.InRoutes.indexOf(newConn.startRoute) * 15) + 5,
                    y: newConn.start.Position.Y - 15
                }, {
                    x: mouse.x,
                    y: mouse.y + 15
                }, {
                    x: mouse.x,
                    y: mouse.y
                }]);
        })
    }

    function terminateConnection() {
        newConnection.style("visibility", "hidden");
        isConnecting = false;
        newConn = {};
    }


    var b = new logReader();
    var c = new uiReader();


});