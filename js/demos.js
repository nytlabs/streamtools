var checkForStreamtools = function() {
  $.ajax({
    dataType: "json",
    url: "http://localhost:7070/status",
    success: function(data) {
      if (data.Blocks && data.Blocks.length > 0) {
        $("#leadPara").after("<p class='text-danger'><span class='glyphicon glyphicon-warning-sign'></span> It looks like you have blocks in your local streamtools. Starting this demo will clear out those blocks. Perhaps you should <a target='_new' href='http://localhost:7070/export'>export before continuing.</p>");
      }

      $("#staticDemo").hide();
      $("#getStarted").show();
    },
    error: function() { $("#getStarted").hide(); $("#staticDemo").show(); }
  });
};

        
var randomElement = function(arr) {
  index = Math.floor(Math.random() * (arr.length - 0) + 0);
  item = arr[index];
  return item;
}
