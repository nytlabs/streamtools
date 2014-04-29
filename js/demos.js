var checkForStreamtools = function() {
  $.ajax({
    dataType: "json",
    url: "http://localhost:7070/status",
    success: function(data) {},
    error: function() { $("#buttonPara").html("<div class='alert alert-warning'><p>Oops! Failed accessing <a target='_new' href='http://localhost:7070'>streamtools</a>. Need to download it? <a href='../#download'>Get the binary</a>, be setup in seconds.</p> <p><a href='" + window.location.href + "'>Reload</a> to see this demo in action once you're ready.</p></div>"); }
  });
};

        
var randomElement = function(arr) {
  index = Math.floor(Math.random() * (arr.length - 0) + 0);
  item = arr[index];
  return item;
}
