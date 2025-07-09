# Developer Guide: Building New GoGallery Templates

This guide explains how to create and structure new templates for GoGallery themes. GoGallery uses Go's `html/template` engine with files ending in `.tmpl.html` for all template pages and partials.

## Template File Structure

A typical theme directory looks like this:

```
themes/
  ThemeName/
    default.tmpl.html         # Base layout template (required)
    pages/
      index.tmpl.html        # Home page
      albums.tmpl.html       # Albums page
      photo.tmpl.html        # Photo detail page
      ...
    partials/
      header.tmpl.html       # Header partial
      footer.tmpl.html       # Footer partial
      ...
    assets/                  # Static assets (css, js, images)
```

## Naming Conventions
- All template files must use the `.tmpl.html` extension.
- Page templates go in the `pages/` directory.
- Reusable partials go in the `partials/` directory.
- The base layout must be named `default.tmpl.html` and reside at the root of the theme.

## Template Syntax
- GoGallery templates use Go's `html/template` syntax:
  - `{{ define "main" }}` ... `{{ end }}` for main content blocks.
  - Use `{{ template "partialName" . }}` to include partials.
  - Access data via dot notation, e.g., `{{.Picture.Name}}`.
- You can use Go template logic: `{{if ...}}`, `{{range ...}}`, etc.

## Adding a New Page Template
1. Create a new file in `pages/`, e.g., `about.tmpl.html`.
2. Start with a `define` block:
   ```gotmpl
   {{ define "main" }}
   <!-- Your HTML and template code here -->
   {{ end }}
   ```
3. Reference the new page in your Go code or navigation as needed.

## Adding/Using Partials
- Place reusable components in `partials/` (e.g., `header.tmpl.html`).
- Include them in your pages or layout with:
  ```gotmpl
  {{ template "header" . }}
  ```

## Data Available in Templates

Below are the most common variables available in templates, depending on the page type:

### Home Page (`index.tmpl.html`)
- `.Albums` — List of all albums (array of Album objects)
- `.FeaturedAlbum` — The featured album (Album object)
- `.Title` — Page title (string)

### Albums Page (`albums.tmpl.html`)
- `.Albums` — List of all albums (array of Album objects)
- `.Title` — Page title (string)

### Collections Page (`collections.tmpl.html`)
- `.Collections` — List of all collections (array of Collection objects)
- `.Title` — Page title (string)

### Photo Page (`photo.tmpl.html`)
- `.Picture` — The current photo (Picture object)
  - `.Name` — Photo name/title
  - `.Caption` — Photo caption/description
  - `.Id` — Photo ID
  - `.Album` — Album ID
  - `.AlbumName` — Album name
  - `.DateTaken` — Date/time photo was taken
  - `.Camera`, `.LensModel`, `.FStop`, `.ShutterSpeed`, `.FocalLength`, `.ISO`, `.ColorSpace`, `.MeteringMode`, `.Software`, `.Saturation`, `.Contrast`, `.Sharpness`, `.Temperature`, `.WhiteBalance` — EXIF/technical fields
- `.PreImagePath` — Path to previous photo (string)
- `.NextImagePath` — Path to next photo (string)

### Pagination Page (`pagination.tmpl.html`)
- `.Items` — List of paginated items (array)
- `.CurrentPage` — Current page number (int)
- `.TotalPages` — Total number of pages (int)

### Common Variables (all pages)
- `.Theme` — Current theme name (string)
- `.BaseURL` — Base URL of the site (string)
- `.User` — Current user (if applicable)

> For a full list of available fields, see the Go structs in the GoGallery codebase (e.g., `Album`, `Picture`, `Collection`).

## Image Usage and Optimization

GoGallery automatically generates multiple image sizes for each photo to optimize loading and display across devices. To ensure the best performance and user experience, follow these guidelines:

### Available Image Sizes
- `xsmall.webp` — Extra small (thumbnails, mobile)
- `small.webp`  — Small (previews, grid views)
- `medium.webp` — Medium (main content, cards)
- `large.webp`  — Large (featured images, banners)
- `xlarge.webp` — Extra large (fullscreen, detail views)

All images are served in modern formats (WebP) for best compression and quality.

### How to Use Responsive Images
Use the `<picture>` element and `srcset` to serve the appropriate image size for each device:

```html
<picture>
  <source srcset="/img/{{.Picture.Id}}/large.webp" type="image/webp">
  <img src="/img/{{.Picture.Id}}/xsmall.webp" alt="{{.Picture.Name}}" loading="lazy" class="w-full h-full object-cover">
</picture>
```
- The browser will pick the best image size based on device and layout.
- Always provide an `alt` attribute for accessibility and SEO.
- Use `loading="lazy"` for images not immediately visible on page load.

### Tips for Template Authors
- Use the smallest image size that looks good for the context (e.g., `xsmall` for thumbnails, `large` or `xlarge` for fullscreen or hero images).
- Use the `<picture>` element for art direction or to provide multiple formats.
- Use CSS classes like `object-cover` or `object-contain` for proper scaling.
- Avoid using original, unoptimized images directly in templates.

### Example: Responsive Gallery Grid
```gotmpl
{{ range .Albums }}
  <a href="/album/{{.Id}}">
    <picture>
      <source srcset="/img/{{.ProfileId}}/small.webp" type="image/webp">
      <img src="/img/{{.ProfileId}}/xsmall.webp" alt="{{.Name}}" loading="lazy" class="rounded shadow object-cover w-full h-32">
    </picture>
    <div>{{.Name}}</div>
  </a>
{{ end }}
```

### Example: Looping Over All Image Sizes

GoGallery provides an `ImgSizes` function in templates, which returns a map of available image sizes. You can use this to generate `<source>` elements for all sizes:

```gotmpl
<picture>
  {{ range $size, $info := ImgSizes }}
    <source srcset="/img/{{.Picture.Id}}/{{$size}}.webp" type="image/webp" media="{{$info.Media}}">
  {{ end }}
  <img src="/img/{{.Picture.Id}}/xsmall.webp" alt="{{.Picture.Name}}" loading="lazy" class="w-full h-full object-cover">
</picture>
```
- `$size` is the size key (e.g., `xsmall`, `small`, `medium`, etc.).
- `$info.Media` is an optional media query string for responsive art direction (if defined in your Go code).
- Always provide a fallback `<img>` tag for browsers that do not support `<source>`.

> See your Go code for the exact structure of `ImgSizes` and available media queries.

By looping over `ImgSizes`, your templates will automatically support all configured image sizes and future changes.

## Tips
- Use Tailwind CSS classes for styling (if your theme supports it).
- Test your templates by running the GoGallery server and navigating to the relevant pages.
- Use Go template comments (`{{/* comment */}}`) to annotate your templates.

## Example: Simple Page Template
```gotmpl
{{ define "main" }}
<h1>{{.Title}}</h1>
<p>Welcome to my custom page!</p>
{{ end }}
```

## Troubleshooting
- If your template does not render, check for syntax errors or missing `define` blocks.
- Ensure all template files use the `.tmpl.html` extension.
- Review the GoGallery logs for error messages.

---
For more advanced usage, see the Go `html/template` documentation: https://pkg.go.dev/html/template
