#!/usr/bin/env perl

use v5.26;
use utf8;
use strict;
use warnings;

$| = 1;

use LWP::Simple;
use HTML::TreeBuilder::XPath;

my $content = get('https://www.raiffeisen.ru/currency_rates/');
die "no content" unless $content;

my $tree= HTML::TreeBuilder::XPath->new;
$tree->parse_content($content);
my $val = $tree->findvalue('//*[@id="online"]/div[2]/div/div/div[2]/div[4]');
die 'no value by xpath //*[@id="online"]/div[2]/div/div/div[2]/div[4]' unless $val;

my $valRub = $val*10266.7;

printf "56.36 * 10000 + 56.53 * 10000 + %.2f * 10266.7 = %s\n", $val, vformat(sprintf "%.2f", 1128900.00 + $valRub);
printf "%.2f * 10266.7 = %s\n", $val, vformat(sprintf "%.2f", $valRub);

exit 0;

sub vformat {
    my $val = shift;
    $val =~ s/(?<=\d)(?=(\d{3})+(?!\d))/ /g;
    return $val;
}
