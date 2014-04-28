var checkForStreamtools = function() {
  $.ajax({
    dataType: "json",
    url: "http://localhost:7070/status",
    success: function(data) {},
    error: function() { $(".jumbotron h1").after("<div class='alert alert-danger'>streamtools doesn't seem to be running! make sure it's accessible at <a href='http://localhost:7070'>http://localhost:7070</a></div>"); }
  });
};

