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

    var width = $(window).width(),
        height = $(window).height();

    var mouse = {
        x: 0,
        y: 0
    };

    var svg = d3.select("body").append("svg")
        .attr("width", width)
        .attr("height", height)
        .on("dblclick", function() {
            d3.event.preventDefault();
            var p = d3.mouse(this);
            $("#create")
                .css({
                    top: p[1],
                    left: p[0],
                    "visibility": "visible"
                });
            $("#create-input").focus();
            mouse.x = p[0];
            mouse.y = p[1];
        });

    $(window).smartresize(function(e) {
        svg.attr("width", $(window).width());
        svg.attr("height", window.innerHeight);
        start();
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
            d3.select(this).attr("transform", function(d, i) {
                return "translate(" + [d.Position.X, d.Position.Y] + ")";
            });
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
                    if (isBlock) {
                        for (var i = 0; i < blocks.length; i++) {
                            blocks.splice(i, 1);
                        }
                    } else {
                        for (var i = 0; i < connections.length; i++) {
                            connections.splice(i, 1);
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
            .call(drag);

        var rects = nodes.append("rect")
            .attr("class", "node");

        var idRects = nodes.append("rect")
            .attr('class', 'idrect');

        nodes.append("svg:text")
            .attr("class", "nodetype")
            .attr("dx", 0)
            .text(function(d) {
                return d.Type;
            }).each(function(d) {
                var bbox = this.getBBox();
                d.width = (d.width > bbox.width ? d.width : bbox.width + 30);
                d.height = (d.height > bbox.height ? d.height : bbox.height + 5);
            }).attr("dy", function(d) {
                return 1 * d.height + 5;
            });

        idRects
            .attr('x', 0)
            .attr('y', 0)
            .attr('width', function(d) {
                return d.width;
            })
            .attr('height', function(d) {
                return d.height * 2;
            });

        node.attr("transform", function(d) {
            return "translate(" + d.Position.X + ", " + d.Position.Y + ")";
        });

        var inRoutes = node.selectAll('.inRoutes')
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
            });

        inRoutes.exit().remove();

        var queryRoutes = node.selectAll('.queryRoutes')
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
            });

        queryRoutes.exit().remove();

        var outRoutes = node.selectAll('.outRoutes')
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
            .attr("class", "rateLabel")
            .attr("startOffset", "50%")
            .attr("xlink:href", function(d) {
                return "#link_" + d.Id;
            })
            .text(function(d) {
                return d.rate;
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

        $.ajax({
            url: '/blocks',
            type: 'POST',
            data: JSON.stringify({
                "Type": blockType,
                "Position": {
                    "X": mouse.x,
                    "Y": mouse.y
                }
            }),
            success: function(result) {}
        });
    }


    var b = new logReader();
    var c = new uiReader();


});