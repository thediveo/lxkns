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

import { Dialog, DialogContent, DialogTitle, IconButton } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import MountpointInfo from 'components/mountpointinfo/MountpointInfo'
import { MountPoint } from 'models/lxkns/mount'
import React, { useContext, useState } from 'react'
import CloseIcon from '@mui/icons-material/Close'
import { ReadonlyIcon } from 'icons/Readonly'
import { NamespaceMap } from 'models/lxkns/model'


const useStyles = makeStyles((theme) => ({
    dialog: {
        marginLeft: 0,
        marginRight: 0,
    },
    close: {
        position: 'absolute',
        right: theme.spacing(1),
        top: theme.spacing(1),
        color: theme.palette.grey[500],
    },
    title: {
        paddingLeft: theme.spacing(2),
        paddingRight: theme.spacing(2),
        '& .MuiSvgIcon-root': {
            position: 'relative',
            verticalAlign: 'baseline',
            top: '0.3ex',
        },
    },
    content: {
        margin: 0,
        paddingLeft: theme.spacing(2),
        paddingRight: theme.spacing(2),
        fontFamily: theme.typography.fontFamily,
        fontSize: theme.typography.body1.fontSize,
    },
}))


const MountpointInfoModalContext = React.createContext<
    null | React.Dispatch<React.SetStateAction<MountPoint>>>(null)


export interface MountpointInfoModalProviderProps {
    /** children to render. */
    children: React.ReactNode
    /** 
     * map of all discovered namespaces for mountpoint namespace root path
     * lookups.
     */
    namespaces: NamespaceMap
}

/**
 * Provider for a MountPoint details modal dialog. Use the setter returned by
 * useMountpointInfoModal() to specify the MountPoint to show the details of and
 * to open the modal details dialog at the same time. Not there is too much
 * "dialog" but rather monologue.
 */
export const MountpointInfoModalProvider = ({
    children,
    namespaces
}: MountpointInfoModalProviderProps) => {

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
                    className={classes.dialog}
                    fullWidth
                    maxWidth={false}
                    scroll="paper"
                    open={!!mountpoint}
                    onClose={handleClose}
                >
                    <DialogTitle className={classes.title}>
                        {mountpoint.mountoptions.includes('ro') && <><ReadonlyIcon fontSize="inherit" />&nbsp;</>}
                        {mountpoint.hidden && 'Hidden '}
                        Mount Point
                        <IconButton
                            aria-label="close"
                            className={classes.close}
                            onClick={handleClose}
                            size="large">
                            <CloseIcon />
                        </IconButton>
                    </DialogTitle>
                    <DialogContent dividers className={classes.content}>
                        <MountpointInfo mountpoint={mountpoint} namespaces={namespaces} />
                    </DialogContent>
                </Dialog>
            }
        </MountpointInfoModalContext.Provider>
    );
}

export default MountpointInfoModalProvider

/**
 * Returns a setter to specify the MountPoint to show information about in a
 * modal dialog.
 */
export const useMountpointInfoModal = () => {
    return useContext(MountpointInfoModalContext)
}
