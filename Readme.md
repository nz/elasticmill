# Elastic Mill

Combine many small Elasticsearch updates into consolidated batch updates.

Elastic Mill behaves as a faux Elasticsearch endpoint for your updates. It proxies all of your read requests, while intercepting, queuing and combining your updates into batches.

## Quick start

ElasticMill has been designed for deployment on Heroku, using the Bonsai Elasticsearch addon.

```bash
git clone http://github.com/nz/elasticmill.git # or use your own fork
cd elasticmill
heroku create --buildpack git://github.com/kr/heroku-buildpack-go.git
heroku addons:add bonsai:starter # or heroku config:add BONSAI_URL=another-app's-bonsai-url
git push heroku master
```

You can now use the web URL from `heroku apps:info` as the Elasticsearch endpoint for your application.

## Example usage

Configure your application to direct all its updates to this server. Individual document updates will be consolidated and batched with other bulk updates and sent to Elasticsearch's Bulk API.

## Why is this necessary?

In general, it is more efficient to update Elasticsearch (and Solr, and Lucene) in fewer consolidated batches, rather than many small and separate updates. This helps consolidate and reduce the overhead from TCP network connections, HTTP request parsing, and JVM garbage collection and cleanup.

However, it can be more efficient and expedient within your application to prepare updates in parallel. Parallelizing is useful to bypass various bottlenecks when loading and serializing your records. It can also be a more natural design pattern to process individual updates to your records as they happen.

The best way to balance these approaches is to use a queue. That way your application gets to leverage the performance benefits of distributing and parallelizing the preparation work, while still consolidating the updates themselves into efficient batch requests for Elasticsearch.

Elasticmill is designed to abstract the implementation of a queue for you. It intercepts many small update requests, and collects and combine them into batches to be sent over to Elasticsearch using its Bulk API.

## Do I need this?

If you're already queuing and asynchronously batching your updates, probably not. If you have a very low volume of updates, probably not. If your Elasticsearch client natively supports this kind of batches, please tell me about it, because I don't know of any who do this natively.

If you have a moderate volume of sustained updates, with peaks measured in updates per second, or if you would benefit from parallelizing your bulk reindexing, this is for you. Direct all of your updates to this server, and it will handle all the batching for you.
