// Copyright 2020 Harald Albrecht.
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

import { useAtom } from 'jotai'

import {
    Box,
    Card,
    Divider,
    Grid,
    List,
    ListItem,
    ListItemText,
    MenuItem,
    Select,
    type SelectChangeEvent,
    styled,
    Switch as Toggle,
    Typography,
} from '@mui/material';
import { 
    expandInitiallyAtom, 
    expandWorkloadInitiallyAtom, 
    showSharedNamespacesAtom, 
    showSystemProcessesAtom, 
    THEME_DARK, 
    THEME_LIGHT, 
    THEME_USERPREF, 
    themeAtom
} from './atoms';


const SettingsGrid = styled(Grid)(({ theme }) => ({
    width: `calc(100% - calc(${theme.spacing(2)} * 2))`,
    margin: theme.spacing(2),

    '& .MuiCard-root + .MuiTypography-subtitle1': {
        marginTop: theme.spacing(4),
    },
}))


/**
 * Renders the "settings" page (view) of the lxkns client browser app.
 */
export const Settings = () => {
    // Tons of settings to play around with...
    const [theme, setTheme] = useAtom(themeAtom)
    const [showSystemProcesses, setShowSystemProcesses] = useAtom(showSystemProcessesAtom)
    const [showSharedNamespaces, setShowSharedNamespaces] = useAtom(showSharedNamespacesAtom)
    const [expandInitially, setExpandInitially] = useAtom(expandInitiallyAtom)
    const [expandWLInitially, setExpandWLInitially] = useAtom(expandWorkloadInitiallyAtom)

    const handleThemeChange = (event: SelectChangeEvent<number>) => {
        setTheme(event.target.value as number)
    }

    return (
        <Box m={1} overflow="auto">
            <SettingsGrid container direction="row" justifyContent="center">
                <Grid
                    container
                    direction="column"
                    style={{ minWidth: '35em', maxWidth: '60em' }}
                >
                    <Typography variant="subtitle1">Appearance</Typography>
                    <Card>
                        <List>
                            <ListItem secondaryAction={
                                <Select value={theme} onChange={handleThemeChange}>
                                    <MenuItem value={THEME_USERPREF}>user preference</MenuItem>
                                    <MenuItem value={THEME_LIGHT}>light</MenuItem>
                                    <MenuItem value={THEME_DARK}>dark</MenuItem>
                                </Select>}
                            >
                                <ListItemText primary="Theme" />
                            </ListItem>
                        </List>
                    </Card>

                    <Typography variant="subtitle1">Display</Typography>
                    <Card>
                        <List>
                            <ListItem secondaryAction={
                                <Toggle
                                    checked={showSystemProcesses}
                                    onChange={() => setShowSystemProcesses(!showSystemProcesses)}
                                    color="primary"
                                />}
                            >
                                <ListItemText
                                    primary="Show system processes"
                                    secondary={(showSystemProcesses ? 'from' : 'hide') + ' /system.slice, /init.scope, /user.slice'}
                                />
                            </ListItem>
                            <Divider />
                            <ListItem secondaryAction={
                                <Toggle
                                    checked={showSharedNamespaces}
                                    onChange={() => setShowSharedNamespaces(!showSharedNamespaces)}
                                    color="primary"
                                />}
                            >
                                <ListItemText
                                    primary="Show shared non-user namespaces"
                                    secondary={showSharedNamespaces
                                        ? 'all namespaces a leader process is attached to'
                                        : 'only namespaces different from parent leader process'}
                                />
                            </ListItem>
                            <Divider />
                            <ListItem secondaryAction={
                                <Toggle
                                    checked={expandInitially}
                                    onChange={() => setExpandInitially(!expandInitially)}
                                    color="primary"
                                />}
                            >
                                <ListItemText
                                    primary="Expand newly discovered namespaces"
                                    secondary={expandInitially
                                        ? 'expand newly discovered namespaces'
                                        : 'expand only top-level new namespaces'}
                                />
                            </ListItem>
                            <Divider />
                            <ListItem secondaryAction={
                                <Toggle
                                    checked={expandWLInitially}
                                    onChange={() => setExpandWLInitially(!expandWLInitially)}
                                    color="primary"
                                />}
                            >
                                <ListItemText
                                    primary="Expand newly discovered containers"
                                    secondary={expandWLInitially
                                        ? 'expand newly discovered containers'
                                        : 'don\'t expand newly discovered containers'}
                                />
                            </ListItem>
                        </List>
                    </Card>

                </Grid>
            </SettingsGrid>
        </Box>
    )
}

export default Settings
