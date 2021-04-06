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

export const filesystemTypeLinks: { [fstype: string]: string } = {
    'autofs': 'https://man7.org/linux/man-pages/man5/autofs.5.html',
    'ext2': 'https://man7.org/linux/man-pages/man5/ext2.5.html',
    'ext3': 'https://man7.org/linux/man-pages/man5/ext3.5.html',
    'ext4': 'https://man7.org/linux/man-pages/man5/ext4.5.html',
    'overlay': 'https://www.kernel.org/doc/html/latest/filesystems/overlayfs.html',
    'proc': 'https://man7.org/linux/man-pages/man5/proc.5.html',
    'squashfs': 'https://www.kernel.org/doc/html/latest/filesystems/squashfs.html',
    'sysfs': 'https://man7.org/linux/man-pages/man5/sysfs.5.html',
    'tmpfs': 'https://man7.org/linux/man-pages/man5/tmpfs.5.html',
    'tracefs': 'https://www.kernel.org/doc/html/latest/trace/ftrace.html#the-file-system',
    'vfat': 'https://www.kernel.org/doc/html/latest/filesystems/vfat.html',
}

export const fallbackFilesystemTypeLink = 'https://man7.org/linux/man-pages/man5/filesystems.5.html'

export const filesystemTypeLink = (fstype: string) =>
    filesystemTypeLinks[fstype] || fallbackFilesystemTypeLink
