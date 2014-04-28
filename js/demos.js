var checkForStreamtools = function() {
  $.ajax({
    dataType: "json",
    url: "http://localhost:7070/status",
    success: function(data) {},
    error: function() { $("#buttonPara").html("<div class='alert alert-warning'><p>Oops! It doesn't look like streamtools is running on your machine.</p><p><a href='https://github.com/nytlabs/streamtools/releases/tag/0.2.5'>Download the latest</a>, start it up and <a href='" + window.location.href + "'>reload this page</a> to see this demo in action.</div>"); }
  });
};

var randomElement = function(arr) {
  index = Math.floor(Math.random() * (arr.length - 0) + 0);
  item = arr[index];
  return item;
}
