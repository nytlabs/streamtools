#! /usr/bin/perl

use strict;
use POSIX qw/strftime/;
use Time::HiRes qw(usleep);
use URI::Escape;
use Date::Parse;
use IO::Socket;

$ENV{'PATH'} = '/bin:/usr/bin';

my $sock = IO::Socket::INET->new(
    Proto    => 'udp',
    PeerPort => 8888,
    PeerAddr => 'localhost',
) or die "Could not create socket: $!\n";

my $message;

my $FMT = "%Y%m%d-%H%M%S";
my $START = str2time("20071209-180000");

my $CURRENT=$START;
my $FINISH=time();

while($CURRENT < $FINISH) {
  my @t = localtime($CURRENT);
  my $YEAR=strftime("%Y", @t);
  my $MONTH=strftime("%m", @t);
  my $TIME=strftime("%M%H%S", @t);
  my $FILENAME=strftime("%Y%m%d-%H%M%S", @t);
  my $TIMESTAMP = strftime("%s", @t);

  # use the line below to direct download the dumps
  my $SOURCE = "http://dumps.wikimedia.org/other/pagecounts-raw/$YEAR/$YEAR-$MONTH/pagecounts-$FILENAME.gz";
  #my $SOURCE = "file:pagecounts-$FILENAME.gz";
  print "Opening source $SOURCE\n";
  open(DUMP, "curl '$SOURCE' 2>/dev/null | /usr/bin/gzip -d -|");
  while(<DUMP>) {
    if(/^([a-zA-Z]+)\.([a-z]+) ([^ ]+) (\d+) (\d+)$/) {
      my $language = $1;
      my $project = $2;
      my $title = uri_unescape($3);
      my $count = $4;

      $title =~ s/[\\\r\n]//g;
      $title =~ s/[\"]/\\"/g;
      $message = "{\"l\":\"$language\",\"p\":\"$project\",\"t\":\"$title\",\"c\":$count,\"ts\":$TIMESTAMP}\n";
    }
    if(/^([a-zA-Z]+) ([^ ]+) (\d+) (\d+)$/) {
      my $language = $1;
      my $title = uri_unescape($2);
      my $count = $3;

      $title =~ s/[\\\r\n]//g;
      $title =~ s/[\"]/\\"/g;
      $message = "{\"l\":\"$language\",\"t\":\"$title\",\"c\":$count,\"ts\":$TIMESTAMP}\n";
    }
    #print "$message";
    $sock->send($message) or die "Send error: $!\n";
    usleep(300);
  }
  close(DUMP);

  exit();
  $CURRENT += 3600;
}