package ncloud

import (
	"fmt"
	"github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/terraform/helper/schema"
	"regexp"
)

func dataSourceNcloudServerProducts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNcloudServerProductsRead,

		Schema: map[string]*schema.Schema{
			"product_name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateRegexp,
			},
			"exclusion_product_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Product code to exclude",
			},
			"product_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Product code to search",
			},
			"server_image_product_code": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Server image product code",
			},
			"region_no": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"zone_no": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"internet_line_type_code": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateInternetLineTypeCode,
				Description:  "Internet line identification code. PUBLC(Public), GLBL(Global). default : PUBLC(Public)",
			},
			"server_products": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"product_code": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"product_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"product_type": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     commonCodeSchemaResource,
						},
						"product_description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"infra_resource_type": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     commonCodeSchemaResource,
						},
						"cpu_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"memory_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"base_block_storage_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"platform_type": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     commonCodeSchemaResource,
						},
						"os_information": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"add_block_storage_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNcloudServerProductsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NcloudSdk).conn

	reqParams := &sdk.RequestGetServerProductList{
		ExclusionProductCode:   d.Get("exclusion_product_code").(string),
		ProductCode:            d.Get("product_code").(string),
		ServerImageProductCode: d.Get("server_image_product_code").(string),
		RegionNo:               d.Get("region_no").(string),
		//ZoneNo:                 d.Get("zone_no").(string),
		//InternetLineTypeCode:   d.Get("internet_line_type_code").(string),
	}

	resp, err := conn.GetServerProductList(reqParams)
	if err != nil {
		logErrorResponse("GetServerProductList", err, reqParams)
		return err
	}
	logCommonResponse("GetServerProductList", reqParams, resp.CommonResponse)

	allServerProducts := resp.Product
	var filteredServerProducts []sdk.Product
	nameRegex, nameRegexOk := d.GetOk("product_name_regex")
	if nameRegexOk {
		r := regexp.MustCompile(nameRegex.(string))
		for _, serverProduct := range allServerProducts {
			if r.MatchString(serverProduct.ProductName) {
				filteredServerProducts = append(filteredServerProducts, serverProduct)
			}
		}
	} else {
		filteredServerProducts = allServerProducts[:]
	}

	if len(filteredServerProducts) < 1 {
		return fmt.Errorf("no results. please change search criteria and try again")
	}

	return serverProductsAttributes(d, filteredServerProducts)
}

func serverProductsAttributes(d *schema.ResourceData, serverImages []sdk.Product) error {
	var ids []string
	var s []map[string]interface{}
	for _, product := range serverImages {
		mapping := map[string]interface{}{
			"product_code": product.ProductCode,
			"product_name": product.ProductName,
			"product_type": map[string]interface{}{
				"code":      product.ProductType.Code,
				"code_name": product.ProductType.CodeName,
			},
			"product_description": product.ProductDescription,
			"infra_resource_type": map[string]interface{}{
				"code":      product.InfraResourceType.Code,
				"code_name": product.InfraResourceType.CodeName,
			},
			"cpu_count":               product.CPUCount,
			"memory_size":             product.MemorySize,
			"base_block_storage_size": product.BaseBlockStorageSize,
			"platform_type": map[string]interface{}{
				"code":      product.PlatformType.Code,
				"code_name": product.PlatformType.CodeName,
			},
			"os_information":         product.OsInformation,
			"add_block_storage_size": product.AddBlockStroageSize,
		}

		ids = append(ids, product.ProductCode)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("server_products", s); err != nil {
		return err
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
