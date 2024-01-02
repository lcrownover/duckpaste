<script>
  import Header from "./Header.svelte";
  import "./styles.css";
</script>

<svelte:head>
  <title>Home</title>
  <meta name="description" content="New Paste" />
</svelte:head>

<div class="app">
  <Header />

  <main>
    <section>
      <h1>
        <span class="new-paste"> New Paste </span>
      </h1>

      <form>
        <div class="row">
          <div class="column">
            <h2>Options</h2>
            <div id="paste-options">
              <input type="radio" id="hours1" name="hours_valid" value="1" />
              <label for="hours1">1 hour</label>
              <input type="radio" id="hours24" name="hours_valid" />
              <label for="24">24 hours</label>
              <input type="radio" id="hours48" name="hours_valid" value="48" />
              <label for="hours48">48 hours</label>
              <input type="radio" id="custom" name="hours_valid" />
              <label for="custom">Custom</label>
              <input
                type="text"
                id="custom-hours"
                name="hours_valid"
                value="8"
              />
              <label for="custom-hours">hours</label>
              <input
                type="radio"
                id="delete-on-read"
                name="hours_valid"
                value="0"
              />
              <label for="delete-on-read">Delete on Read</label>
              <button>Save</button>
            </div>
          </div>
          <div class="column">
            <div id="paste-input">
              <textarea rows="32"></textarea>
            </div>
          </div>
        </div>
      </form>
    </section>
  </main>
</div>
