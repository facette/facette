<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-modal name="chart-preview" :title="i18n.t('labels.charts.preview')" @show="onShow">
        <template v-slot="modal">
            <v-form>
                <v-chart tooltip v-model:value="chart"></v-chart>

                <template v-slot:bottom>
                    <v-button @click="modal.close(false)">
                        {{ i18n.t("labels.close") }}
                    </v-button>

                    <v-button primary @click="createChart" v-autofocus>
                        {{ i18n.t("labels.charts.create") }}
                    </v-button>
                </template>
            </v-form>
        </template>
    </v-modal>
</template>

<script lang="ts">
import cloneDeep from "lodash/cloneDeep";
import {ref} from "vue";
import {useI18n} from "vue-i18n";
import {useRouter} from "vue-router";
import {useStore} from "vuex";

import {Chart} from "types/api";

import {State} from "@/store";

export interface ModalChartPreviewParams {
    expr: string;
}

const defaultChart: Chart = {
    id: "",
    name: "",
    options: {
        type: "area",
    },
    series: [],
};

export default {
    setup(): Record<string, unknown> {
        const i18n = useI18n();
        const router = useRouter();
        const store = useStore<State>();

        const chart = ref<Chart>(cloneDeep(defaultChart));

        const createChart = (): void => {
            // Prefill series data into store
            store.commit("routeData", {prefill: chart.value.series});

            router.push({name: "admin-charts-edit", params: {id: "new"}});
        };

        const onShow = (params: ModalChartPreviewParams): void => {
            chart.value.series = [{expr: params.expr}];
        };

        return {
            chart,
            createChart,
            i18n,
            onShow,
        };
    },
};
</script>

<style lang="scss" scoped>
.v-modal {
    ::v-deep(.v-modal-content) {
        display: flex;
        flex-direction: column;
        height: 80vh;
        width: 80vw;

        .v-form {
            flex-grow: 1;
            overflow: hidden;

            .v-chart {
                height: calc(100% - 3rem - var(--button-height));
            }
        }
    }
}
</style>
