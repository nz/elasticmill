# Elastic Mill _(alpha)_

Combine many small Elasticsearch updates into consolidated batch updates.

Elastic Mill behaves as a faux Elasticsearch endpoint for your updates. It proxies all of your read requests, while intercepting, queuing and combining your updates into batches.

**Experimental work in progress:** If you feel adventurous, give it a try, it seems to work. Let me know how it goes. Otherwise, give me a few more days to test and polish.

## Quick start

ElasticMill has been designed for deployment on Heroku, using the Bonsai Elasticsearch addon — my tools of choice. But it should run pretty much anywhere.

```bash
git clone http://github.com/nz/elasticmill.git # or use your own fork
cd elasticmill
git checkout v1-wip
heroku create --buildpack git://github.com/kr/heroku-buildpack-go.git
heroku addons:add bonsai:starter # or heroku config:add BONSAI_URL=...
git push heroku HEAD:master
heroku run 'curl -X POST $BONSAI_URL/test' # create a test index
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

## Contributing

There are a handful of useful enhancements to be made here.

- *Tests.* Erm, yeah, this thing needs tests. If you find a bug, I'd love a test that duplicates it.
- *Logging.* This proxy is perfectly positioned to log all sorts of useful metrics.
- *External Queuing.* The first version is using Go channels for simplicity, effectively queuing in memory. But that presents some mild complications when scaling beyond a single process. It would be nice to use a more robust external queue, such as [IronMQ](http://iron.io/mq) built by my friends at [Iron.io](http://iron.io/)
- *Solr* support. All of these ideas work equally well for Apache Solr.
- *Caching* and *rate limiting.* Could be useful!

To contribute:

1. Fork the repo.
2. Add a test for your change.
3. Optionally implement a fix for that change.
4. Push your commits in a named remote branch.
5. Open a pull request, and optionally [email Nick](mailto:nick@bonsai.io) with a link.

## Who built this?

Elastic Mill was built by [Nick Zadrozny](http://nick.zadrozny.com), one of the co-founders of One More Cloud, and a co-creator of the [Bonsai Hosted Elasticsearch](http://www.bonsai.io/) service and [Websolr](https://websolr.com/). Nick does search all day every day.

Its creation was prompted while helping a customer optimize their application's reindexing performance to get the best balance between parallelizing and batching. The design was born out of years of frustration at the difficulty of integrating queuing into existing clients, and a minor epiphany that clients need not care about such implementation details.

