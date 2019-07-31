# Batch operator examples

Automi can batch streaming elements so that
certain operations can be applied.  This directory
contains examples of Automi batch operators and how they are used.

* [groupby_key](./groupby_key) - Shows how to batch and group streaming map values based on their keys.
* [groupby_name](./groupby_name) - Batches and group struct streaming values based on field names.
* [groupby_pos](./groupby_pos) - Batches and group slice/array streaming values based on index position.
* [sort](./sort) - Sorts batched items using the natural sort sequence of streamed items.
* [sort_with](./sort_with) - Shows how to sort items using a custom sorting function.
* [sort_key](./sort_key) - Shows how to sort streaming map values based on their keys.
* [sort_name](./sort_name) - Shows how to sort streaming struct values based on field names.
* [sort_pos](./sort_pos) - Shows how to sort streaming slice values based on selected index.
* [sum](./sum) - Example that shows how to add streaming numeric values.
* [sumby_key](./sumby_key) - Shows how to add streaming map values based on their keys.
* [sumby_name](./sumby_name) - Shows how to add streaming struct values basd on field names.
* [sumby_pos](./sumby_pos) - Example that shows how to add straming slice values based on selected index.