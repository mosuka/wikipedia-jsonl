# wikipedia-jsonl

wikipedia-jsonl is a CLI that converts Wikipedia dump XML to JSON Lines format.


## Requirement

This command uses [SQLite](https://sqlite.org). Make sure to install SQLite for your platform in advance.


## Download Wikipedia dumps

Download Wikipedia dumps from [Wikimedia Downloads](https://dumps.wikimedia.org/backup-index.html).

- enwiki-YYYYMMDD-pages-articles-multistream.xml.bz2
- enwiki-YYYYMMDD-categorylinks.sql.gz


## Import dumps

Checkout [mysql2sqlite](https://github.com/dumblob/mysql2sqlite)

```
% git clone git@github.com:dumblob/mysql2sqlite.git
```

Convert the Dump file to Sqlite SQL and import it into Sqlite.

```
% gunzip -c enwiki-20211201-categorylinks.sql.gz | ./mysql2sqlite/mysql2sqlite - | sqlite3 enwiki-20211201.db
```

## Convert Wikipedia XML to JSONL

Run the following command to convert the XML to JSONL and output it to stdout.

```
% bzcat enwiki-20211201-pages-articles-multistream.xml.bz2 | ./bin/wikipedia-jsonl -a -c -d enwiki-20211201.db -e -m -l -r
```

Executing the above command will output the results as shown below.

```
{"categories":["Redirects_from_moves","Redirects_with_old_history","Unprintworthy_redirects"],"external_links":[],"id":10,"links":[{"Namespace":"","PageName":"Computer accessibility","Anchor":""}],"media":[],"redirect":"Computer accessibility","text":" Computer accessibility","timestamp":"2021-01-23T15:15:01Z","title":"AccessibleComputing"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"external_links":[],"id":14,"links":[{"Namespace":"","PageName":"Geography of Afghanistan","Anchor":""}],"media":[],"redirect":"Geography of Afghanistan","text":" Geography of Afghanistan","timestamp":"2017-06-05T04:18:23Z","title":"AfghanistanGeography"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"external_links":[],"id":15,"links":[{"Namespace":"","PageName":"Demographics of Afghanistan","Anchor":""}],"media":[],"redirect":"Demographics of Afghanistan","text":" Demographics of Afghanistan","timestamp":"2017-06-05T04:19:42Z","title":"AfghanistanPeople"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"external_links":[],"id":18,"links":[{"Namespace":"","PageName":"Communications in Afghanistan","Anchor":""}],"media":[],"redirect":"Communications in Afghanistan","text":" Communications in Afghanistan","timestamp":"2017-06-05T04:19:45Z","title":"AfghanistanCommunications"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"external_links":[],"id":19,"links":[{"Namespace":"","PageName":"Transport in Afghanistan","Anchor":""}],"media":[],"redirect":"Transport in Afghanistan","text":" Transport in Afghanistan","timestamp":"2017-06-04T21:42:11Z","title":"AfghanistanTransportations"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"external_links":[],"id":20,"links":[{"Namespace":"","PageName":"Afghan Armed Forces","Anchor":""}],"media":[],"redirect":"Afghan Armed Forces","text":" Afghan Armed Forces","timestamp":"2017-06-04T21:43:11Z","title":"AfghanistanMilitary"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"external_links":[],"id":21,"links":[{"Namespace":"","PageName":"Foreign relations of Afghanistan","Anchor":""}],"media":[],"redirect":"Foreign relations of Afghanistan","text":" Foreign relations of Afghanistan","timestamp":"2017-06-04T21:43:14Z","title":"AfghanistanTransnationalIssues"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"external_links":[],"id":23,"links":[{"Namespace":"","PageName":"Assistive technology","Anchor":""}],"media":[],"redirect":"Assistive technology","text":" Assistive_technology","timestamp":"2017-06-05T04:19:50Z","title":"AssistiveTechnology"}

...
```

