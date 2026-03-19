// Copyright 2026 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

import { Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, styled, Typography } from "@mui/material"
import clsx from "clsx"
import { useRef, useState, type ChangeEvent } from "react"

const DropZone = styled(Box)(({ theme }) => ({
    display: 'block',
    borderWidth: '2px',
    borderStyle: 'dashed',
    borderColor: theme.palette.divider,
    borderRadius: '2px',
    margin: theme.spacing(0),
    padding: theme.spacing(2),
    textAlign: 'center',

    '&.draggedover': {
        borderStyle: 'solid',
        borderColor: theme.palette.primary.main,
    }
}))

export interface DiscoveryUploaderProps {
    /** show or hide the modal dialog */
    open: boolean
    /** handler to be called when the dialog is closed for whichever reason */
    onClose?: () => void
    /** handler to be called with contents when a file has been dropped or selected */
    onImport?: (content: string) => void
}

/**
 * The `DiscoveryUploader` component renders a model dialog that allows
 * selecting a file or dropping a file, preferably with correct JSON discovery
 * contents.
 */
export const DiscoveryUploader = ({ open, onClose, onImport }: DiscoveryUploaderProps) => {

    const [file, setFile] = useState<File | null>(null)
    const [inDrag, setInDrag] = useState(false)

    // we need a reference to the (hidden) file selector input field in order to
    // trigger the dialog by simulating a click on the input field.
    const inputRef = useRef<HTMLInputElement | null>(null)

    // when closing the dialog, reset our state so we're clean for the next
    // show.
    const handleClose = () => {
        onClose?.()
        setFile(null)
        setInDrag(false)
    }

    // when asked to import, asynchronously get the dropped file's contents and
    // call the parent-supplied callback handler with the concrete string contents.
    const handleImport = async () => {
        if (onImport && file) {
            const content = await file.text()
            onImport?.(content)
        }
        handleClose()
    }

    // when the user selected a file, remember this choice so we can show the
    // file name in the drop zone as feedback as well as a reminder to the user.
    const handleInputChange = (event: ChangeEvent<HTMLInputElement>) => {
        const f = event.target.files?.[0] ?? null
        if (f) {
            setFile(f)
        }
    }

    // give visual feedback during drag&drop within the drop zone; if the drag
    // doesn't include files (like text selections, URLs, HTML content, ...)
    // then do not give any positive visual feedback.
    const handleDragOver = (event: React.DragEvent<HTMLDivElement>) => {
        event.preventDefault()
        if (Array.from(event.dataTransfer.types).includes("Files")) {
            setInDrag(true)
        }
    }

    // give visual feedback during drag&drop within the drop zone.
    const handleDragLeave = (event: React.DragEvent<HTMLDivElement>) => {
        event.preventDefault()
        setInDrag(false)
    }

    // give visual feedback during drag&drop within the drop zone; additionally,
    // on drop pick up the first file (if any) and remember it as the user's
    // choice in order to give visual feedback and a reminder to the user.
    const handleDrop = (event: React.DragEvent<HTMLDivElement>) => {
        event.preventDefault()
        setInDrag(false)
        const f = event.dataTransfer.files?.[0] ?? null
        if (f && (f.type === 'application/json' || f.name.toLowerCase().endsWith('.json'))) {
            setFile(f)
        }
    }

    return <Dialog
        open={open}
        onClose={handleClose}
        slotProps={{
            paper: {
                sx: { backgroundImage: 'none' }
            }
        }}
    >
        <DialogTitle>Import Discovery Data</DialogTitle>
        <DialogContent>
            <DropZone
                className={clsx(inDrag && 'draggedover')}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
            >
                <Typography variant="body1" gutterBottom>
                    Drag & drop a discovery JSON data file here, or click to browse.
                </Typography>
                <Button
                    variant="contained"
                    onClick={() => inputRef.current?.click()}
                    sx={{ mt: 2 }}
                >
                    Browse
                </Button>
                {file &&
                    <Typography
                        variant="body2"
                        color="text.secondary"
                        sx={{ mt: 2 }}
                    >
                        {file.name}
                    </Typography>
                }
            </DropZone>
            <input
                hidden
                ref={inputRef}
                type="file"
                accept=".json,application/json"
                onChange={handleInputChange}
            />
        </DialogContent>
        <DialogActions>
            <Button onClick={handleClose} color="primary">
                Cancel
            </Button>
            <Button onClick={handleImport} disabled={!file} color="primary">
                Import
            </Button>
        </DialogActions>
    </Dialog>
}
