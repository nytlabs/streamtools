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
    //var HOST = "http://localhost:7080/";
    var HOST = "";
    var width = $(window).width(),
        height = $(window).height();

    var color = d3.scale.category10();

    var nodes = [],
        links = [],
        labels;

    var force = d3.layout.force()
        .nodes(nodes)
        .links(links)
        .charge(-1000)
        .linkDistance(100)
        .linkStrength(0.5)
        .size([width, height])
        .on("tick", tick);

    var svg = d3.select("body").append("svg")
        .attr("width", width)
        .attr("height", height);

    var linkContainer = svg.append('g')
        .attr('class', 'linkContainer');

    var nodeContainer = svg.append('g')
        .attr('class', 'nodeContainer');

    var link = svg.select(".linkContainer").selectAll(".link"),
        node = svg.select(".nodeContainer").selectAll(".node");

    svg.append("svg:defs").selectAll("marker")
        .data(["arrow"])
        .enter().append("svg:marker")
        .attr("id", String)
        .attr("viewBox", "0 0 20 20")
        .attr("refX", 20)
        .attr("refY", 10)
        .attr("markerWidth", 10)
        .attr("markerHeight", 10)
        .attr("orient", "auto")
        .append("svg:path")
        .attr("d", "M 0 0 L 20 10 L 0 20 z");

    var blocks = {};
    var connections = {};

    function update() {
        $.get(HOST + "list", function(data) {

            var newIDs = [];

            data.forEach(function(b) {
                newIDs.push(b.ID);
            });

            for (var key in connections) {
                if (newIDs.indexOf(key) === -1) {
                    for (var i = 0; i < links.length; i++) {
                        if (links[i].id == key) {
                            links.splice(i, 1);
                            delete connections[key];
                            break;
                        }
                    }
                }
            }

            for (var key in blocks) {
                if (newIDs.indexOf(key) === -1) {
                    for (var i = 0; i < nodes.length; i++) {
                        if (nodes[i].id == key) {
                            nodes.splice(i, 1);
                            delete blocks[key];
                            break;
                        }
                    }
                }
            }

            for (var i = 0; i < data.length; i++) {
                if (data[i].BlockType != "connection" && !blocks.hasOwnProperty(data[i].ID)) {
                    data[i].id = data[i].ID;
                    data[i].width = 0;
                    data[i].height = 0;
                    data[i].hsl = getHSL(data[i].BlockType);
                    blocks[data[i].ID] = data[i];
                    nodes.push(blocks[data[i].ID]);
                }
            }

            for (var i = 0; i < data.length; i++) {
                if (data[i].BlockType == "connection") {
                    var connID = data[i].ID;
                    if (!connections.hasOwnProperty(connID)) {
                        connections[connID] = data[i];
                        links.push({
                            hsl: getHSL(data[i].BlockType),
                            id: data[i].ID,
                            rate: 0,
                            rateLoc: 0,
                            BlockType: data[i].BlockType,
                            source: blocks[data[i].InBlocks[0]],
                            target: blocks[data[i].OutBlocks[0]]
                        });
                    }
                }
            }

            console.log("updated");
            start();
        });
    }

    update();

    $.get(HOST + "library", function(data) {

        data = data.sort(function(a, b) {
            if (a.BlockType > b.BlockType) return 1;
            if (a.BlockType < b.BlockType) return -1;
            return 0;
        });

        for (var i = 0; i < data.length; i++) {
            if (data[i].BlockType != "connection") {
                $("#library")
                    .append($("<option></option>")
                        .attr("value", data[i].BlockType)
                        .text(data[i].BlockType));
            }
        }

        $('#create').on('click', function() {
            var id = window.prompt('create an id for this block', '');
            $.get('http://localhost:7080/create?blockType=' + $('#library').val() + '&id=' + id, function(data) {
                update();
            });
        });
    });

    $(window).smartresize(function(e) {
        svg.attr("width", $(window).width());
        svg.attr("height", window.innerHeight);
        force.size([$(window).width(), window.innerHeight]);
        start();
    });

    $('#block_bar').hide();

    var CONNECT_ZERO_STATE = 0;
    var CONNECT_SOURCE_STATE = 1;
    var CONNECT_TARGET_STATE = 2;
    var connectSource = null;
    var connectState = CONNECT_ZERO_STATE;

    $('#connect').on('click', function(d) {
        connectState = 1;
        $('#source').html('select a source block');
    });

    function start() {
        link = link.data(force.links(), function(d) {
            return d.id;
        });
        link.enter()
            .append("line", ".node")
            .attr("class", "link");
        link.exit().remove();

        label = svg.selectAll('.edgeRate')
            .data(force.links(), function(d) {
                return d.id;
            });

        label.enter().append('text')
            .attr("class", "edgeRate")
            .attr("x", function(d) {
                return (d.source.y + d.target.y) / 2;
            })
            .attr("y", function(d) {
                return (d.source.x + d.target.x) / 2;
            })
            .attr("text-anchor", "middle")
            .text(function(d) {
                return 0;
            })
            .on("mousedown", function(d) {
                var infoTmp = $('#block_info').html();
                var connTmp = $('#conn_rule').html();

                $.get(HOST + "blocks/" + d.id + "/last_message", function(data) {
                    var tmpl = _.template(connTmp, {
                        data: data
                    });
                    $("#rule").html(tmpl);
                });

                $("#info").html(_.template(infoTmp, {
                    block: d
                }));

                $("#delete").unbind().on('click', function() {
                    $.get('http://localhost:7080/delete?id=' + d.id, function(data) {
                        update();
                    });
                });
            });

        label.exit().remove();

        var ping = svg.select('.linkContainer').selectAll(".edgePing")
            .data(force.links(), function(d) {
                return d.id;
            });

        ping.enter().append("circle")
            .attr("class", "edgePing")
            .attr("r", 4);

        ping.exit().remove();

        node = node.data(force.nodes(), function(d) {
            return d.id;
        });

        var nodes = node.enter()
            .append("g")
            .call(force.drag);

        nodes.on("mousedown", function(d) {
            if (connectState == 2) {
                connectState = 0;
                console.log('http://localhost:7080/connect?from=' + connectSource + '&to=' + d.id);
                $.get('http://localhost:7080/connect?from=' + connectSource + '&to=' + d.id, function(data) {
                    update();
                });
            }

            if (connectState == 1) {
                connectSource = d.id;
                connectState = 2;
                $('#source').html('source: ' + d.id);
                $('#target').html('select a target block');
            }


            $('#block_bar').show();

            var infoTmp = $('#block_info').html();
            var ruleTmp = $('#block_rule').html();

            if (d.Routes.indexOf("get_rule") !== -1) {
                $.get(HOST + "blocks/" + d.ID + "/get_rule", function(ruleData) {
                    var tmpl = _.template(ruleTmp, {
                        block: d,
                        rule: ruleData,
                        routes: d.Routes
                    });
                    $("#rule").html(tmpl);

                    $("#update").on("click", function() {
                        var rule = {};
                        for (var key in ruleData) {
                            var ruleInput = $('#' + d.ID + "_" + key);
                            var val = ruleInput.val();
                            var type = ruleInput.prop("tagName");

                            switch (typeof(ruleData[key])) {
                                case 'boolean':
                                    rule[key] = val === 'true' ? true : false;
                                    break;
                                case 'string':
                                    rule[key] = val;
                                    break;
                                case 'object':
                                    rule[key] = JSON.parse(val);
                                    break;
                                case 'number':
                                    rule[key] = parseFloat(val);
                                    break;
                            }
                        }
                        $.post(HOST + "blocks/" + d.ID + "/set_rule", JSON.stringify(rule), function(data) {
                            //console.log(data)
                        });
                    });
                });
            } else {
                $("#rule").empty();
            }
            $("#info").html(_.template(infoTmp, {
                block: d
            }));

            $("#delete").unbind().on('click', function() {
                //console.log('http://localhost:7080/delete?id=' + d.ID)
                $.get('http://localhost:7080/delete?id=' + d.ID, function(data) {
                    update();
                    $('#block_bar').hide();
                });
            });

        });

        nodes.on("mouseover", function(d) {
            d3.select(this).attr('class', '');
        });

        var rects = nodes.append("rect")
            .attr("class", "node")
        //.attr("width", 50)
        //.attr("height", 40)
        .attr("fill", function(d) {
            return getHSL(d.BlockType);
        });

        nodes.append("svg:text")
            .attr("class", "nodetext")
            .attr("dx", 0)
            .attr("dy", 0)
            .text(function(d) {
                return d.BlockType;
            }).each(function(d) {
                var bbox = this.getBBox();
                d.width = d.width > bbox.width ? d.width : bbox.width;
                d.height = d.height > bbox.height ? d.height : bbox.height;
            }).attr("dy", function(d) {
                return 1 * d.height;
            })

        nodes.append("svg:text")
            .attr("class", "nodetext")
            .attr("dx", 0)
            .attr("dy", function(d) {
                return 2 * d.height;
            })
            .text(function(d) {
                return d.ID;
            }).each(function(d) {
                var bbox = this.getBBox();
                d.width = d.width > bbox.width ? d.width : bbox.width;
                d.height = d.height > bbox.height ? d.height : bbox.height;
            });

        rects
            .attr('width', function(d) {
                return d.width;
            })
            .attr('height', function(d) {
                return d.height * 2;
            })


        node.exit().remove();

        force.start();
    }

    function getHSL(str) {
        var strLen = str.length;
        var d = Math.floor(strLen / 3);
        var strmax = d * 3;
        var hsl = [];

        for (var i = 0; i < strmax; i += d) {
            var c = 0;
            for (var j = i; j < i + d; j++) {
                c += str.charCodeAt(j) - 64;
            }
            hsl.push(Math.floor(c /= d));
        }
        return 'hsl(' + (20 * (hsl[0] + hsl[1] + hsl[2])) + ',' + 100 + '%,' + 50 + '%)';
    }

    function tick() {
        node.attr("transform", function(d) {
            return "translate(" + (d.x - (d.width * .5)) + ", " + (d.y - (d.height)) + ")";
        });

        link.each(function(d) {
            var s1 = d.source;
            var s2 = d.target;
            var width = d.target.width;
            var height = d.target.height * 2;
            var x2 = s2.x;
            var y2 = s2.y;
            var intersection = intersect_line_box(s1, s2, {
                x: x2 - width / 2.0,
                y: y2 - height / 2.0
            }, width, height);

            if (!intersection) {
                d.intersection = {
                    x: s2.x,
                    y: s2.y
                };
            }

            d.intersection = {
                x: intersection.x,
                y: intersection.y
            };
        })
            .attr("x1", function(d) {
                return d.source.x;
            })
            .attr("y1", function(d) {
                return d.source.y;
            })
            .attr("x2", function(d) {
                return d.intersection.x;
            })
            .attr("y2", function(d) {
                return d.intersection.y;
            })
            .attr("class", "link arrow")
            .attr("marker-end", "url(#arrow)");

        svg.selectAll('.edgeRate')
            .attr("x", function(d) {
                return (d.source.x + d.target.x) / 2;
            })
            .attr("y", function(d) {
                return 6 + (d.source.y + d.target.y) / 2;
            });
    }

    window.setInterval(function() {
        svg.selectAll('.edgePing')
            .each(function(d) {
                d.rateLoc += 0.001 + Math.min(d.rate, 100) / 4000.0;
                if (d.rateLoc > .75) d.rateLoc = 0;
            })
            .attr('cx', function(d) {
                return d.source.x + (d.target.x - d.source.x) * d.rateLoc;
            })
            .attr('cy', function(d) {
                return d.source.y + (d.target.y - d.source.y) * d.rateLoc;
            });
    }, 1000 / 60);

    //update edge rates
    window.setInterval(function() {
        svg.selectAll('.edgeRate')
            .each(function(d) {
                $.get(HOST + "blocks/" + d.id + "/rate", function(data) {
                    d.rate = Math.round(100 * data.rate) / 100.0;
                });
            });

        svg.selectAll('.edgeRate')
            .text(function(d) {
                return d.rate;
            });
    }, 1000);


    var cursor = document.getElementById('cursor');
    var termStart = document.getElementById('start');
    var end = document.getElementById('end');

    var cursorInverted = true;
    var cmdHistory = [''];
    var cmdHistoryIndex = 0;


    var _to_ascii = {
        '188': '44',
        '109': '45',
        '190': '46',
        '191': '47',
        '192': '96',
        '220': '92',
        '222': '39',
        '221': '93',
        '219': '91',
        '173': '45',
        '187': '61', //IE Key codes
        '186': '59', //IE Key codes
        '189': '45' //IE Key codes
    };

    var shiftUps = {
        "96": "~",
        "49": "!",
        "50": "@",
        "51": "#",
        "52": "$",
        "53": "%",
        "54": "^",
        "55": "&",
        "56": "*",
        "57": "(",
        "48": ")",
        "45": "_",
        "61": "+",
        "91": "{",
        "93": "}",
        "92": "|",
        "59": ":",
        "39": "\"",
        "44": "<",
        "46": ">",
        "47": "?"
    };

    window.setInterval(function() {
        cursorInverted = !cursorInverted;
        if (cursorInverted) {
            cursor.classList.add('inverted');
        } else {
            cursor.classList.remove('inverted');
        }
    }, 500);

    function safe(str) {
        return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/ /g, '&nbsp;');
    }

    function response(str) {
        var r = document.getElementById('respTxt');
        r.innerHTML = syntaxHighlight(str);
    }

    function clearResponse() {
        var r = document.getElementById('respTxt');
        r.innerHTML = '';
    }

    var rootCommands = {
        create: function(tokens) {
            if (tokens.length > 3) {
                console.log("incorrect number of arguments");
                return false;
            }
            var query = HOST + "create?blockType=" + tokens[1];
            if (tokens.length == 3) {
                query += '&id=' + tokens[2];
            }

            $.ajax({
                url: query,
                success: function(data) {
                    update();
                    console.log(data);
                    response(data);
                },
                statusCode: {
                    500: function(e) {
                        response(JSON.parse(e.responseText));
                    }
                }
            });
            return true;
        },
        delete: function(tokens) {
            if (tokens.length > 2) {
                console.log("incorrect number of arguments");
                return false;
            }
            var query = HOST + "delete?id=" + tokens[1];

            $.ajax({
                url: query,
                success: function(data) {
                    update();
                    console.log(data);
                    response(data);
                },
                statusCode: {
                    500: function(e) {
                        response(JSON.parse(e.responseText));
                    }
                }
            });
            return true;
        },
        connect: function(tokens) {
            if (tokens.length !== 3) {
                console.log("incorrect number of arguments");
                return false;
            }
            var query = HOST + "connect?from=" + tokens[1] + "&to=" + tokens[2];
            $.ajax({
                url: query,
                success: function(data) {
                    update();
                    console.log(data);
                    response(data);
                },
                statusCode: {
                    500: function(e) {
                        response(JSON.parse(e.responseText));
                    }
                }
            });
            return true;
        },
    };

    function routeCommand(tokens, obj) {
        var routeTokens = tokens[0].split('/');
        var query = 'http://localhost:7080/blocks/' + routeTokens[0] + '/' + routeTokens[1];

        if (obj === null) {
            $.ajax({
                url: query,
                success: function(data) {
                    update();
                    console.log(data);
                    response(data);
                },
                statusCode: {
                    500: function(e) {
                        response(JSON.parse(e.responseText));
                    }
                }
            });
        } else {
            var sendObj;
            try {
                sendObj = JSON.parse(obj);
            } catch (e) {
                response(JSON.parse('{"error":"invalid json"}'));
                return false;
            }
            $.post(query, JSON.stringify(sendObj), function(data) {
                response(data);
            });
        }

        return true;
    }

    function execute(cmd) {
        var obj = null;

        if (cmd.indexOf('{') !== -1) {
            obj = cmd.substring(cmd.indexOf('{'));
            cmd = cmd.substring(0, cmd.indexOf('{'));
        }

        var tokens = cmd.replace(/\s+/g, ' ').trim().split(' ');

        if (tokens === null || tokens.length === 0) {
            return;
        }

        if (rootCommands.hasOwnProperty(tokens[0]) && rootCommands[tokens[0]](tokens)) {
            //update();
            return;
        }

        if (tokens[0].indexOf('/') !== -1 && routeCommand(tokens, obj)) {
            //update();
            return;
        }

        response(JSON.parse('{"error":"invalid command"}'));
        return;
    }

    var termOn = false;
    $('#term').hide();
    $('#hiddenTerm').focusout(function(e) {
        termOn = false;
        $('#term').hide();
    });

    document.onkeydown = function(e) {
        if (e.which == 192) {
            e.preventDefault();
            e.stopPropagation();

            termOn = !termOn;
            if (termOn) {
                $('#term').show();
                $('#hiddenTerm').focus();
            } else {
                $('#hiddenTerm').blur();
                $('#term').hide();
            }
            return;
        }
    };

    document.onkeyup = function(e) {
        if (e.which == 8 || e.which == 46) {
            e.preventDefault();
            e.stopPropagation();
        }

        if (e.which == 192) {
            return;
        }

        if (termOn === false) return;

        var cursorIndex = $('#hiddenTerm')[0].selectionStart;
        var cmd = $('#hiddenTerm').val();

        if (e.which == 8 || e.which == 46) {
            e.preventDefault();
            e.stopPropagation();
        }

        if (e.which == 38) {
            // up
            if (cmdHistoryIndex > 0) {
                cmdHistoryIndex--;
                cmd = cmdHistory[cmdHistoryIndex];
                $('#hiddenTerm').val(cmd);
                cursorIndex = cmd.length;
            }
        } else if (e.which == 40) {
            // down
            if (cmdHistoryIndex < cmdHistory.length - 1) {
                cmdHistoryIndex++;
                cmd = cmdHistory[cmdHistoryIndex];
                $('#hiddenTerm').val(cmd);
                cursorIndex = cmd.length;
            }
        } else if (e.which == 13) {
            e.preventDefault();
            e.stopPropagation();
            if (cmd.length === 0) {
                clearResponse();
            } else {
                if (cmdHistoryIndex != cmdHistory.length - 1) {
                    cmdHistory[cmdHistory.length - 1] = cmd;
                }

                cmdHistory.push(cmd);

                cmdHistoryIndex = cmdHistory.length - 1;
                execute(cmd);
                cmd = '';
                $('#hiddenTerm').val('');
                cmdHistory[cmdHistory.length - 1] = '';
                cursorIndex = 0;
            }
        } else if (e.which !== 16) {
            if (cmdHistoryIndex == cmdHistory.length - 1) {
                cmdHistory[cmdHistory.length - 1] = cmd;
            }
        }

        termStart.innerHTML = safe(cmd.substring(0, cursorIndex));

        if (cursorIndex === cmd.length) {
            cursor.innerHTML = '&nbsp;';
        } else {
            cursor.innerHTML = safe(cmd.substring(cursorIndex, cursorIndex + 1));
        }

        end.innerHTML = safe(cmd.substring(cursorIndex + 1, cmd.length));
    };

    //http://stackoverflow.com/questions/4810841/json-pretty-print-using-javascript
    function syntaxHighlight(json) {
        json = JSON.stringify(json, undefined, 4);
        json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
        return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function(match) {
            var cls = 'number';
            if (/^"/.test(match)) {
                if (/:$/.test(match)) {
                    cls = 'key';
                } else {
                    cls = 'string';
                }
            } else if (/true|false/.test(match)) {
                cls = 'boolean';
            } else if (/null/.test(match)) {
                cls = 'null';
            }
            return '<span class="' + cls + '">' + match + '</span>';
        });
    }

    // from Springy.js
    // https://github.com/dhotson/springy/blob/master/springyui.js
    function intersect_line_box(p1, p2, p3, w, h) {
        var tl = {
            x: p3.x,
            y: p3.y
        };
        var tr = {
            x: p3.x + w,
            y: p3.y
        };
        var bl = {
            x: p3.x,
            y: p3.y + h
        };
        var br = {
            x: p3.x + w,
            y: p3.y + h
        };

        var result;
        if (result = intersect_line_line(p1, p2, tl, tr)) {
            return result;
        } // top
        if (result = intersect_line_line(p1, p2, tr, br)) {
            return result;
        } // right
        if (result = intersect_line_line(p1, p2, br, bl)) {
            return result;
        } // bottom
        if (result = intersect_line_line(p1, p2, bl, tl)) {
            return result;
        } // left

        return false;
    }

    function intersect_line_line(p1, p2, p3, p4) {
        var denom = ((p4.y - p3.y) * (p2.x - p1.x) - (p4.x - p3.x) * (p2.y - p1.y));

        // lines are parallel
        if (denom === 0) {
            return false;
        }

        var ua = ((p4.x - p3.x) * (p1.y - p3.y) - (p4.y - p3.y) * (p1.x - p3.x)) / denom;
        var ub = ((p2.x - p1.x) * (p1.y - p3.y) - (p2.y - p1.y) * (p1.x - p3.x)) / denom;

        if (ua < 0 || ua > 1 || ub < 0 || ub > 1) {
            return false;
        }

        return {
            x: p1.x + ua * (p2.x - p1.x),
            y: p1.y + ua * (p2.y - p1.y)
        };
    }
});