# Download and Import Discoveries

![downloading and importing discoveries](_images/lxkns-appbar-impexport.png ':class=framedscreenshot')

- ➊ download discovery data as JSON,
- ➋ import the discovery data at any later point in time.

## Downloading Discoveries

After the first successful discovery, tapping or clicking the action bar button
➊ will download the currently shown discovery in form of a JSON file. This is
the same JSON that the **lxkns** backend responds with on the `/api/namespaces`
REST API endpoint.

## Importing Discoveries

At any later time you can revisit downloaded discoveries by importing them so
you can view them. Tap or click the action bar button ➋. Next, an import dialog
will be shown.

![importing discoveries dialog](_images/import-dialog.png ':class=framedscreenshot')

You can either drag and drop a JSON file of a downloaded discovery into the
marked zone inside the dialog, or you can press the "BROWSE" button to select a
discovery file.

Next, press the "IMPORT" button. This will close the dialog and the imported
discovery is shown until the next refresh. Please note that importing a
discovery will automatically switch off any automatic refresh.
