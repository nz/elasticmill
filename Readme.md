# Elastic Mill

Combine many small Elasticsearch updates into consolidated batch updates.

Elastic Mill behaves as a faux Elasticsearch endpoint for your updates. It consists of two components:

1. The _server,_ which intercepts all update requests and inserts the document JSON into a message queue for later processing.
2. The _processor,_ which pulls document JSON out of the message queue, and sends them to Elasticearch in consolidated bulk updates.

## Quick start

ElasticMill has been designed for deployment on Heroku, using the Bonsai Elasticsearch addons.

```bash
git clone http://github.com/onemorecloud/elasticmill.git
cd elasticmill
heroku create --buildpack git://github.com/kr/heroku-buildpack-go.git
heroku addons:add iron_mq:developer
heroku addons:add bonsai:starter # or heroku config:add BONSAI_URL=another-app's-bonsai-url
git push heroku master
```

You can now use the web URL from `heroku apps:info` as an Elasticsearch endpoint for updates.

## Example usage

Configure your application to direct all its updates to this server. Individual document updates will be consolidated and batched with other bulk updates and sent to Elasticsearch's Bulk API.

## Why is this necessary?

In general, it is more efficient to update Elasticsearch (and Solr, and Lucene) in fewer consolidated batches, rather than many small and separate updates. This helps consolidate and reduce the overhead from TCP network connections, HTTP request parsing, and JVM garbage collection and cleanup.

However, it can be more efficient and expedient to prepare updates from your application in a highly parallel manner. Parallelizing is useful to bypass bottlenecks in loading your records from your database and serializing them into JSON to be indexed by Elasticsearch. It is also a more natural design pattern to process individual updates to your records as they happen.

The best way to balance these approaches is to use a queue. That way your application gets to leverage the performance benefits of distributing and parallelizing the work of preparing updates for Elasticsearch, while still consolidating the updates themselves into efficient batch requests.

Elasticmill is designed to abstract the implementation of a queue for you. Its server intercepts many small update requests, and inserts them into a queue. Its processor continually pulls updates from the queue and sends them to Elasticsearch in efficient batches.

## Do I need this?

If you're already queuing and asynchronously batching your updates, probably not. If you have a very low volume of updates, probably not. If your Elasticsearch client natively supports this kind of batches, please tell me about it, because I don't know of any who do this natively.

If you have a moderate volume of sustained updates, with peaks measured in updates per second, or if you would benefit from parallelizing your bulk reindexing, this is for you. Direct all of your updates to this server, and it will handle all the queuing and batching for you.

