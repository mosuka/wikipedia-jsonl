# wikipedia-jsonl

wikipedia-jsonl is a CLI that converts Wikipedia dump XML to JSON Lines format.


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
% bzcat enwiki-20211201-pages-articles-multistream.xml.bz2 | ./bin/wikipedia-jsonl -a -c -d enwiki-20211201.db
```

Executing the above command will output the results as shown below.

```
{"categories":["Redirects_from_moves","Redirects_with_old_history","Unprintworthy_redirects"],"id":10,"text":" Computer accessibility\n\n\n\n","timestamp":"2021-01-23T15:15:01Z","title":"AccessibleComputing"}
{"categories":["All_articles_lacking_reliable_references","Anarchism","Anti-capitalism","Anti-fascism","Articles_containing_French-language_text","Articles_containing_Spanish-language_text","Articles_lacking_page_references_from_October_2021","Articles_lacking_reliable_references_from_August_2021","Articles_prone_to_spam_from_November_2014","Articles_with_BNE_identifiers","Articles_with_BNF_identifiers","Articles_with_EMU_identifiers","Articles_with_GND_identifiers","Articles_with_HDS_identifiers","Articles_with_LCCN_identifiers","Articles_with_NKC_identifiers","Articles_with_short_description","CS1:_long_volume_value","Economic_ideologies","Good_articles","Left-wing_politics","Libertarian_socialism","Libertarianism","Political_culture","Political_ideologies","Political_movements","Short_description_matches_Wikidata","Social_theories","Socialism","Use_British_English_from_August_2021","Use_dmy_dates_from_August_2021","Wikipedia_articles_needing_page_number_citations_from_October_2021","Wikipedia_indefinitely_semi-protected_pages"],"id":12,"text":"\n\n\n\n\n\n\n\nAnarchism is a political philosophy and movement that is sceptical of authority and rejects all involuntary, coercive forms of hierarchy. Anarchism calls for the abolition of the state, which it holds to be unnecessary, undesirable, and harmful. As a historically left-wing movement, placed on the farthest left of the political spectrum, it is usually described alongside libertarian Marxism as the libertarian wing (libertarian socialism) of the socialist movement, and has a strong historical association with anti-capitalism and socialism.\n\n\nHumans lived in societies without formal hierarchies long before the establishment of formal states, realms, or empires. With the rise of organised hierarchical bodies, scepticism toward authority also rose; in the 19th century, a self-conscious political movement emerged questioning all forms of authority. During the latter half of the 19th and the first decades of the 20th century, the anarchist movement flourished in most parts of the world and had a significant role in workers' struggles for emancipation. Various anarchist schools of thought formed during this period. Anarchists have taken part in several revolutions, most notably in the Comune of Paris, Russian Civil War and Spanish Civil War, whose end marked the end of the classical era of anarchism. In the last decades of the 20th and into the 21st century, the anarchist movement has been resurgent once more.\n\n\nAnarchism employs a diversity of tactics in order to meet its ideal ends which can be broadly separated into revolutionary and evolutionary tactics; there is significant overlap between the two, which are merely descriptive. Revolutionary tactics aim to bring down authority and state, having taken a violent turn in the past, while evolutionary tactics aim to prefigure what an anarchist society would be like. Anarchist thought, criticism, and praxis have played a part in diverse areas of human society. Criticism of anarchism include claims that it is internally inconsistent, violent, or utopian.\n\n\n\n\n","timestamp":"2021-11-28T13:17:21Z","title":"Anarchism"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"id":13,"text":" History of Afghanistan\n\n\n\n","timestamp":"2017-06-05T04:18:18Z","title":"AfghanistanHistory"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"id":14,"text":" Geography of Afghanistan\n\n\n\n","timestamp":"2017-06-05T04:18:23Z","title":"AfghanistanGeography"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"id":15,"text":" Demographics of Afghanistan\n\n\n\n","timestamp":"2017-06-05T04:19:42Z","title":"AfghanistanPeople"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"id":18,"text":" Communications in Afghanistan\n\n\n\n","timestamp":"2017-06-05T04:19:45Z","title":"AfghanistanCommunications"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"id":19,"text":" Transport in Afghanistan\n\n\n\n","timestamp":"2017-06-04T21:42:11Z","title":"AfghanistanTransportations"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"id":20,"text":" Afghan Armed Forces\n\n\n\n","timestamp":"2017-06-04T21:43:11Z","title":"AfghanistanMilitary"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"id":21,"text":" Foreign relations of Afghanistan\n\n\n\n","timestamp":"2017-06-04T21:43:14Z","title":"AfghanistanTransnationalIssues"}
{"categories":["Redirects_with_old_history","Unprintworthy_redirects"],"id":23,"text":" Assistive_technology\n\n\n\n","timestamp":"2017-06-05T04:19:50Z","title":"AssistiveTechnology"}

...
```

