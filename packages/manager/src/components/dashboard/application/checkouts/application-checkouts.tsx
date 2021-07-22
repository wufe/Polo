import { IApplication, IApplicationBranchModel } from '@polo/common/state/models';
import { observer } from 'mobx-react-lite';
import React, { useState } from 'react';
import { ApplicationCheckout } from '../checkout/application-checkout';

type TProps = {
    branches                   : IApplication['branchesMap'];
    tags                       : IApplication['tagsMap'];
    onSessionCreationSubmission: (checkout: string) => void;
}

export const ApplicationCheckouts = observer((props: TProps) => {
    const [selectBranches, setSelectBranches] = useState(true);

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