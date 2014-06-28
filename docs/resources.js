$(function() {
    var results;
    $.get("databib.csv", function(data) {
        var databib;
        databib = $.parse(data);

        $.each(databib.results.fields, function(k, v) {
            $(".info-row").append("<th>" + v + "</th>");
            return (v !== "Description");
        });        

        var tableBody;
        $.each(databib.results.rows, function(index, item) {
            tableBody += "<tr>"
            $.each(item, function(key, prop) {
                if (key == "URL") {
                    tableBody += "<td><a target='_new' href='" + prop + "'>" + prop + "</a>";
                } else {
                    tableBody += "<td>" + prop + "</td>";
                }
                return (key !== "Description");
            }); 
            tableBody += "</tr>"
        });
        $("table").append(tableBody);
    });
});
