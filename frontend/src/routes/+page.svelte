<svelte:head>
  <title>Home</title>
  <meta name="description" content="Svelte demo app" />
</svelte:head>

<section>
  <h1>
    <span class="new-paste"> New Paste </span>
  </h1>

  <form>
    <textarea rows="10"></textarea>
    <input type="radio" id="hours1" name="hours1" value="hours1" /><span
      id="hours1">1 hour</span
    >
    <input type="radio" id="hours24" name="hours24" value="hours24" /><span
      id="hours24">24 hours</span
    >
    <input type="radio" id="hours48" name="hours48" value="hours48" /><span
      id="hours48">48 hours</span
    >
    <input type="radio" id="custom" name="custom" value="custom" /><span
      id="custom">Custom</span
    >
    <input type="text" id="custom-hours" name="custom-hours" value="8" /><span
      id="custom-hours">hours</span
    >
    <input
      type="radio"
      id="delete-on-read"
      name="delete-on-read"
      value="delete-on-read"
    /><span id="delete-on-read">Delete on Read</span>
    <button>Save</button>
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

  h1 {
    width: 100%;
  }

  .new-paste {
    display: block;
    position: relative;
    width: 100%;
    height: 100%;
  }

  form {
    display: flex;
    flex-direction: column;
    width: 100%;
    max-width: 64rem;
    margin: 0 auto;
    height: 64rem;
  }

  textarea {
    height: auto;
  }
</style>
