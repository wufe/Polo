import { IApplication, IApplicationBranchModel } from '@/state/models';
import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useState } from 'react';
import dayjs from 'dayjs';
import { ApplicationCheckout } from '../checkout/application-checkout';

type TProps = {
    branches                   : IApplication['branchesMap'];
    tags                       : IApplication['tagsMap'];
    onSessionCreationSubmission: (checkout: string) => void;
}

export const ApplicationCheckouts = observer((props: TProps) => {

    const [openBranches, setOpenBranches] = useState<{
        [k:string]: boolean;
    }>({});
    const [selectBranches, setSelectBranches] = useState(true);

    const toggleBranch = (name: string) => setOpenBranches(b => ({...b, [name]: !b[name]}));

    const checkoutsToShow = selectBranches ?
        sortBranches(Array.from(props.branches.values())) :
        Array.from(props.tags.values());

    return <>
        <h4 className="my-1 text-lg">New session</h4>
        <span className="text-sm text-gray-500 opacity-80">
            Build a new session by choosing a build point.
        </span>
        <div className="flex justify-center mt-4 mb-3">
            <div className="border border-gray-400 dark:border-gray-600 rounded-md overflow-hidden inline-flex flex-nowrap items-stretch text-xs">
                <div
                    onClick={() => setSelectBranches(true)}
                    className={`flex flex-nowrap items-center px-3 py-2 cursor-pointer hover:bg-nord4 dark:hover:bg-nord10
                        ${selectBranches ? 'bg-nord4 dark:bg-nord10' : ''}`}>
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        className="w-4 h-4 mr-1"
                        viewBox="0 0 512 512"
                        fill="currentColor">
                        <path d="M416 160a64 64 0 10-96.27 55.24c-2.29 29.08-20.08 37-75 48.42-17.76 3.68-35.93 7.45-52.71 13.93v-126.2a64 64 0 10-64 0v209.22a64 64 0 1064.42.24c2.39-18 16-24.33 65.26-34.52 27.43-5.67 55.78-11.54 79.78-26.95 29-18.58 44.53-46.78 46.36-83.89A64 64 0 00416 160zM160 64a32 32 0 11-32 32 32 32 0 0132-32zm0 384a32 32 0 1132-32 32 32 0 01-32 32zm192-256a32 32 0 1132-32 32 32 0 01-32 32z"></path>
                    </svg>
                    <span>Branches</span>
                </div>
                <div className="border-r border-gray-500 dark:border-gray-600" style={{ width: 1 }}></div>
                <div
                    onClick={() => setSelectBranches(false)}
                    className={`flex flex-nowrap items-center px-3 py-2 cursor-pointer hover:bg-nord4 dark:hover:bg-nord10
                        ${selectBranches ? '' : 'bg-nord4 dark:bg-nord10'}`}>
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        className="w-4 h-4 mr-1"
                        viewBox="0 0 24 24"
                        stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                    </svg>
                    <span>Tags</span>
                </div>
            </div>
        </div>

        <div className="divide-y dark:divide-gray-700">
            {checkoutsToShow.map((checkout, key) =>
                <ApplicationCheckout
                    key={key}
                    type={selectBranches ? 'branch' : 'tag'}
                    onSessionCreationSubmission={props.onSessionCreationSubmission}
                    {...checkout} />
                // <div
                //     className="flex flex-col" key={key}>
                //     <div className="flex items-end lg:items-center pt-1 pb-2 px-2 lg:px-6 cursor-pointer lg:dark:hover:bg-nord-1" onClick={() => toggleBranch(branch.name)}>
                //         <div className="flex-1 min-w-0">
                //             <div className="text-xs text-gray-500 uppercase hidden lg:block">Name</div>
                //             <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis text-center lg:text-left py-5 lg:py-0" title={branch.name}>{branch.name}</div>
                //         </div>
                //         <span className="text-center col-span-2 hidden lg:inline-block">
                //             <span className="text-sm underline cursor-pointer inline-block mx-3 hover:text-blue-400" onClick={() => props.onSessionCreationSubmission(branch.name)}>Create session</span>
                //         </span>
                //     </div>
                //     <div className={`col-span-12 ${!openBranches[branch.name] && 'hidden'}`}>
                //         <div className="grid grid-cols-12 gap-2 items-center p-3 lg:px-6 bg-nord5 dark:bg-nord1 rounded-sm">
                //             <div className="col-span-12">
                //                 <div className="text-xs text-gray-500 uppercase">Message</div>
                //                 <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis">{branch.message}</div>
                //             </div>
                //             <div className="col-span-5">
                //                 <div className="text-xs text-gray-500 uppercase">Author</div>
                //                 <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis">{branch.author}</div>
                //             </div>
                //             <div className="col-span-7">
                //                 <div className="text-xs text-gray-500 uppercase">Date</div>
                //                 <div className="text-sm whitespace-nowrap overflow-hidden overflow-ellipsis">{dayjs(branch.date).format('DD MMM HH:mm')}</div>
                //             </div>
                //             <div className="py-2 lg:hidden text-center col-span-12">
                //                 <span className="leading-none text-sm underline cursor-pointer inline-block mx-3 hover:text-nord14 border border-nord14 rounded-md py-2 px-10" onClick={() => props.onSessionCreationSubmission(branch.name)}>Create session</span>
                //             </div>
                //         </div>
                        
                //     </div>
                    
                // </div>
                )}
        </div>
        
    </>;
});

function sortBranches(branches: IApplicationBranchModel[]): IApplicationBranchModel[] {
    const preferred = [
        'master',
        'main',
        'hotfix',
        'develop',
        'dev',
        'feature'
    ];

    let result: IApplicationBranchModel[] = [];
    let length = branches.length;

    for (const pref of preferred) {
        for (let i = 0; i < length; i++) {
            const branch = branches[i];
            if (branch.name.toLowerCase().startsWith(pref.toLowerCase())) {
                result.push(branch);
                branches.splice(i, 1)
                length--;
                i--;
            }
        }
    }

    return result.concat(branches);
}