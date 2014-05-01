var checkForStreamtools = function() {
  $.ajax({
    dataType: "json",
    url: "http://localhost:7070/status",
    success: function(data) { $("#staticDemo").hide(); $("#getStarted").show(); },
    error: function() { $("#getStarted").hide(); $("#staticDemo").show(); }
  });
};

        
var randomElement = function(arr) {
  index = Math.floor(Math.random() * (arr.length - 0) + 0);
  item = arr[index];
  return item;
}
