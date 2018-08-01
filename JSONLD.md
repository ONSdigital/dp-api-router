JSON-LD Files
=====================

The JSON-LD files live in the [cdn assets](https://github.com/ONSdigital/cdn.ons.gov.uk-assets). These files currently provide the JSON-LD view of the Dataset API. In future, by adding new fields to these files, this could be extended to cover other APIs as well.

### How to update

If new fields are added to the Dataset API endpoints, or if a new API is to be mapped, pull out every data field name and see if that concept is already mapped. If it is not, add a line in the `context.json` file to map that field to a relevant vocabulary definition. These definitions could be from any vocabulary, but will likely come from:

Dublin Core: http://purl.org/dc/terms/,
DCAT: http://www.w3.org/ns/dcat#,
Schema.org: https://schema.org,
Stat-DCAT: http://data.europa.eu/(xyz)/statdcat-ap/,
Linked Data Cube: http://purl.org/linked-data/cube#,

If no reasonable definiton can be found in any linked data vocabulary, consider if this is a niche ONS term. In this case add a stanza to the `terms.json` file.

Once each of these files has been updated as needed, pull request and merge in the CSN repo as normal. There is currently no expectation that multiple versions or history of the JSON context is provided to users. This will load the updated file to the CDN in the same location, meaning changes will take more r less immediate effect.

If the context has been updated to cover another API which has not previously had context, update the [API Router](https://github.com/ONSdigital/dp-api-router) as per this [PR](https://github.com/ONSdigital/dp-api-router/pull/31) so the `NewAPIProxy` for the relevant API contains the config item referencing the `context.json` location in the CDN.
