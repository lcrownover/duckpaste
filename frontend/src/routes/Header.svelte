<script>
  import { page } from "$app/stores";
  import logo from "$lib/images/UOSignature-WHT.png";
  import github from "$lib/images/github-white.png";
</script>

<header>
  <div class="corner corner-left">
    <a href="https://uoregon.edu">
      <img src={logo} alt="University of Oregon" />
    </a>
  </div>

  <nav>
    <ul>
      <li aria-current={$page.url.pathname === "/" ? "page" : undefined}>
        <a href="/">Home</a>
      </li>
      <li aria-current={$page.url.pathname === "/about" ? "page" : undefined}>
        <a href="/about">About</a>
      </li>
    </ul>
  </nav>

  <div class="corner corner-right">
    <a
      href="https://github.com/lcrownover/duckpaste"
      target="_blank"
      rel="noopener noreferrer"
    >
      <img src={github} alt="GitHub" />
    </a>
  </div>
</header>
