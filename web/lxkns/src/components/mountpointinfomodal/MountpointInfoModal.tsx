// Copyright 2021 Harald Albrecht.
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

import { Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton, makeStyles } from '@material-ui/core'
import MountpointInfo from 'components/mountpointinfo/MountpointInfo'
import { MountPoint } from 'models/lxkns/mount'
import React, { useContext, useState } from 'react'
import CloseIcon from '@material-ui/icons/Close'


const useStyles = makeStyles((theme) => ({
    close: {
        position: 'absolute',
        right: theme.spacing(1),
        top: theme.spacing(1),
        color: theme.palette.grey[500],
    },
    content: {
        fontFamily: theme.typography.fontFamily,
        fontSize: theme.typography.body1.fontSize,
        paddingLeft: 0,
        margin: `0 ${theme.spacing(2)}px`,
    },
}))


const MountpointInfoModalContext = React.createContext<
    null | React.Dispatch<React.SetStateAction<MountPoint>>>(null)


export interface MountpointInfoModalProviderProps {
    /** children to render. */
    children: React.ReactNode
}

/**
 * Provider for a MountPoint details modal dialog. Use the setter returned by
 * useMountpointInfoModal() to specify the MountPoint to show the details of and
 * to open the modal details dialog at the same time. Not there is too much
 * "dialog" but rather monologue.
 */
export const MountpointInfoModalProvider = ({ children }: MountpointInfoModalProviderProps) => {

    const classes = useStyles()

    const [mountpoint, setMountpoint] = useState(null as MountPoint)

    const handleClose = () => {
        setMountpoint(null)
    }

    return (
        <MountpointInfoModalContext.Provider value={setMountpoint}>
            {children}
            {mountpoint &&
                <Dialog
                    fullWidth
                    maxWidth={false}
                    scroll="paper"
                    open={!!mountpoint}
                    onClose={handleClose}
                >
                    <DialogTitle>
                        {mountpoint.hidden && 'Hidden '}
                        Mount Point
                        <IconButton
                            aria-label="close"
                            className={classes.close}
                            onClick={handleClose}
                        >
                            <CloseIcon />
                        </IconButton>
                    </DialogTitle>
                    <DialogContent dividers className={classes.content}>
                        <MountpointInfo mountpoint={mountpoint} />
                    </DialogContent>
                    <DialogActions>
                        <Button autoFocus onClick={handleClose} color="primary">
                            Close
                        </Button>
                    </DialogActions>
                </Dialog>
            }
        </MountpointInfoModalContext.Provider>
    )
}

export default MountpointInfoModalProvider

/**
 * Returns a setter to specify the MountPoint to show information about in a
 * modal dialog.
 */
export const useMountpointInfoModal = () => {
    return useContext(MountpointInfoModalContext)
}
