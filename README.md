# wikipedia-jsonl

wikipedia-jsonl is a CLI that converts Wikipedia dump XML to JSON Lines format.


## How to use

The following command will convert the XML to JSONL and output them to the standard output.

```
% bzcat enwiki-20211201-pages-articles-multistream.xml.bz2 | wikipedia-jsonl
```

Executing the above command will output the results as shown below.

```
{"id":10,"title":"AccessibleComputing","text":"Computer accessibility","timestamp":"2021-01-23T15:15:01Z","ns":0,"redirect":"Computer accessibility"}
{"id":13,"title":"AfghanistanHistory","text":"History of Afghanistan","timestamp":"2017-06-05T04:18:18Z","ns":0,"redirect":"History of Afghanistan"}
{"id":14,"title":"AfghanistanGeography","text":"Geography of Afghanistan","timestamp":"2017-06-05T04:18:23Z","ns":0,"redirect":"Geography of Afghanistan"}
{"id":15,"title":"AfghanistanPeople","text":"Demographics of Afghanistan","timestamp":"2017-06-05T04:19:42Z","ns":0,"redirect":"Demographics of Afghanistan"}

...
```
