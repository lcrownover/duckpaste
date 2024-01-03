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
      <form>
        <div class="row">
          <div class="column">
            <h2>Options</h2>

            <div class="paste-options">
              <div class="options">
                <label for="options">paste expiration</label>
                <select class="select-options" name="paste-options">
                  <option value="48">1 Week</option>
                  <option value="48">48 Hours</option>
                  <option value="24">24 Hours</option>
                  <option value="8">8 Hours</option>
                  <option value="1">1 Hour</option>
                </select>
              </div>

              <div class="options">
                <label for="burn">delete upon reading</label>
                <input type="checkbox" class="burn" name="burn" />
              </div>

              <div class="submit">
                <button>Save</button>
              </div>
            </div>
          </div>
          <div class="column">
            <div id="paste-input">
              <textarea cols="80" rows="32"></textarea>
            </div>
          </div>
        </div>
      </form>
    </section>
  </main>
</div>
