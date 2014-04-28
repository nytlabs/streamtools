$(document).ready(function() {
  $(".btn-demos").click(function(e) {
    e.preventDefault;
    window.document.location = "/demos";
  });
  $(".btn-download").click(function(e) {
    e.preventDefault;
    window.document.location = "https://github.com/nytlabs/streamtools/releases";
  });
  $(".btn-learn").click(function(e) {
    e.preventDefault;
    window.document.location = "https://github.com/nytlabs/streamtools";
  });
});
