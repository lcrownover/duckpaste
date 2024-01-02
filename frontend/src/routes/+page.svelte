<svelte:head>
  <title>Home</title>
  <meta name="description" content="New Paste" />
</svelte:head>

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
          <input type="text" id="custom-hours" name="hours_valid" value="8" />
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

<style>
  section {
    display: flex;
    flex-direction: column;
    justify-content: left;
    align-items: left;
    flex: 0.6;
  }

  .row {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    width: 100%;
  }

  .column {
    display: flex;
    flex-direction: column;
    flex-basis: 100%;
    flex: 1;
  }

  #paste-input {
    padding: 2rem;
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100%;
  }

  #paste-options {
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100%;
  }

  h1 {
    width: 100%;
  }

  .new-paste {
    display: block;
    position: relative;
    width: 100%;
    height: 100%;
  }

  /* form {
    display: flex;
    flex-direction: column;
    width: 100%;
    max-width: 64rem;
    margin: 0 auto;
    height: 64rem;
  } */

  textarea {
    height: auto;
    white-space: pre;
    overflow-wrap: normal;
    overflow-x: scroll;
  }
</style>
