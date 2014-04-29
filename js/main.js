$(document).ready(function(){

    // Display correct binary to download
    var OSName="";
    if (navigator.appVersion.indexOf("Win")!=-1) OSName="Windows";
    if (navigator.appVersion.indexOf("Mac")!=-1) OSName="MacOS";
    if (navigator.appVersion.indexOf("X11")!=-1) OSName="UNIX";
    if (navigator.appVersion.indexOf("Linux")!=-1) OSName="Linux";

    if (OSName == "Windows")    $(".download-windows").addClass('selected-OS');
    if (OSName == "MacOS")      $(".download-mac").addClass('selected-OS');
    if (OSName == "Linux")      $(".download-linux").addClass('selected-OS');
    if (OSName == "")           $(".download-none").addClass('selected-OS')

    $(".download").not(".selected-OS").css('display', 'none');

});