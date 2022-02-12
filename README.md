# wikipedia-jsonl

wikipedia-jsonl is a CLI that converts Wikipedia dump XML to JSON Lines format.


## Pretaration

Download Wikipedia dumps from [Wikimedia Downloads](https://dumps.wikimedia.org/backup-index.html).

- enwiki-YYYYMMDD-pages-articles-multistream.xml.bz2
- enwiki-YYYYMMDD-categorylinks.sql.gz
- enwiki-YYYYMMDD-page.sql.gz


## Import dumps

onvert the Dump file to Sqlite SQL and import it into Sqlite.

```
% gunzip -c enwiki-20211201-page.sql.gz | ./bin/mysql2sqlite - | sqlite3 enwiki-20211201.db
% gunzip -c enwiki-20211201-categorylinks.sql.gz | ./bin/mysql2sqlite - | sqlite3 enwiki-20211201.db
```

## Convert Wikipedia XML to JSONL

Run the following command to convert the XML to JSONL and output it to stdout.

```
% bzcat enwiki-20211201-pages-articles-multistream.xml.bz2 | wikipedia-jsonl --abstruct
```

Executing the above command will output the results as shown below.

```
{"id":10,"title":"AccessibleComputing","text":"Computer accessibility","timestamp":"2021-01-23T15:15:01Z","ns":0,"redirect":"Computer accessibility"}
{"id":13,"title":"AfghanistanHistory","text":"History of Afghanistan","timestamp":"2017-06-05T04:18:18Z","ns":0,"redirect":"History of Afghanistan"}
{"id":14,"title":"AfghanistanGeography","text":"Geography of Afghanistan","timestamp":"2017-06-05T04:18:23Z","ns":0,"redirect":"Geography of Afghanistan"}
{"id":15,"title":"AfghanistanPeople","text":"Demographics of Afghanistan","timestamp":"2017-06-05T04:19:42Z","ns":0,"redirect":"Demographics of Afghanistan"}

...
```

