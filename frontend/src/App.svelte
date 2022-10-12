<script lang="ts">
  // class QueryResult {
  //   path: string;
  //   title: string;
  //   offsets: string;
  //   snippet: string;
  //
  //   static createFrom(source: any = {}) {
  //     return new QueryResult(source);
  //   }
  //
  //   constructor(source: any = {}) {
  //     if ('string' === typeof source) source = JSON.parse(source);
  //     this.path = source["path"];
  //     this.title = source["title"];
  //     this.offsets = source["offsets"];
  //     this.snippet = source["snippet"];
  //   }
  // }
  //
  import {Open, Query} from '../wailsjs/go/main/App'
  // import {main} from "../wailsjs/go/models";
  // import QueryResult = main.QueryResult;

  let input_area: string = ""
  let input_value: string = ""

  let results = []
  let selected = ""
  let snippet = ""

  function keydown(e) {
    if (e.key !== "Enter") {
      return
    }
    Open(selected)
  }

  $: input_value = input_area
  $: (async () => {
    results = await Query(input_value)
    selected = ""
    snippet = ""
  })()
  $: {
    console.log(results.length)
    console.log(selected)
    if (results.length > 0 && selected != "") {
      snippet = results.find(e => e.path == selected).snippet
    }
  }
</script>

<p>hello</p>

<p>input_value: {input_value}</p>

<!--<p>length: {length}</p>-->

<ul>
  <!--{#each results as result}-->
  <!--  <li>{result.path}></li>-->
  <!--{/each}-->
</ul>

<div id="container">
  <div id="itemA">
    Query: <input type="text" bind:value={input_area} spellcheck="false" autofocus>
  </div>
  <div id="itemB">
    <select size="20" bind:value={selected} on:keydown={keydown}>
      {#each results as result}
        <option value={result.path}>{result.path}</option>
      {/each}
    </select>
  </div>
  <div id="itemC">
    <p>selected: {selected}</p>
    <p>snippet: {@html snippet}</p>
  </div>
</div>
