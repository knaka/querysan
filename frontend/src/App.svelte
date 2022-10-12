<script lang="ts">
  import {Body, Open, Query} from '../wailsjs/go/main/App'

  let input_text: string = ""
  let query: string = ""
  // 日本語入力を伴うと、text input へ bind された変数へ
  // 入ってくる文字列の内容が危ういので、都度コピーしてみたら
  // うまく行った。なぜかは分からない
  $: query = input_text

  let query_results = []
  let selected_path = ""
  let snippet = ""
  let body = ""
  let seq = -1

  function keydown(e) {
    if (e.key !== "Enter") { return }
    Open(selected_path)
  }

  function double_click(e) {
    Open(e.target.value)
  }

  function input_keydown(e) {
    if (e.isComposing) {
      return
    }
    console.log(e)
    query = e.target.value
    // input_text
  }

  $: (async () => {
    console.log("query:", query, seq)
    let result_new = await Query(query, seq ++)
    if (result_new["error"]) {
      return
    }
    query_results = result_new["results"]
    // if (query_results.length > 0) {
    //   selected_path = query_results[0]["path"]
    // } else {
    selected_path = ""
    snippet = ""
    body = ""
    // }
  })()

  $: (async () => {
    body = await Body(selected_path)
  })()

  $: if (query_results.length > 0 && selected_path != "") {
    snippet = query_results.find(e => e["path"] == selected_path)["snippet"]
  }
</script>

<style>
  #wrapper {
    width: 100%;
    height: 100%;
    margin: 0;
    display: grid;
    grid-template-columns: 20% 80%;
    grid-template-rows: 100%;
  }

  #left_pane {
    margin: 0;
    grid-column: 1 / 2;
    grid-row: 1 / 2;
  }

  #query_pane {
    margin: 0;
  }

  #selection_pane {
    margin: 0;
  }

  #selection_pane select {
    height: 100%;
  }

  #selection {
    width: 100%;
    height: 100%;
  }

  #itemC {
    margin: 1ex;
    grid-column: 2 / 3;
    grid-row: 1 / 2;
  }

  pre {
    white-space: pre-wrap ;
  }
</style>

<div id="wrapper">
  <div id="left_pane">
    <div id="query_pane">
      <!-- Get rid of "A11y: Avoid using autofocus" warning · Issue #6629 · sveltejs/svelte https://github.com/sveltejs/svelte/issues/6629 -->
      <!-- svelte-ignore a11y-autofocus -->
      Query: <input type="text" spellcheck="false" id="text_input" autofocus on:keydown={input_keydown}>
    </div>
    <div id="selection_pane">
      <select size=30 id=selection bind:value={selected_path} on:keydown={keydown} on:dblclick={double_click}>
        {#each query_results as result}
          <option value={result["path"]}>{result["path"]}</option>
        {/each}
      </select>
    </div>
  </div>
  <div id="itemC">
    <p>{@html snippet}</p>
    <hr>
    <pre>{@html body}</pre>
  </div>
</div>
