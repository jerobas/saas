export namespace main {
	
	export class InventoryBatchDTO {
	    id: string;
	    item_id: string;
	    quantity_total: number;
	    quantity_remaining: number;
	    purchase_price_total: number;
	    unit_price: number;
	    purchased_at: string;
	
	    static createFrom(source: any = {}) {
	        return new InventoryBatchDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.item_id = source["item_id"];
	        this.quantity_total = source["quantity_total"];
	        this.quantity_remaining = source["quantity_remaining"];
	        this.purchase_price_total = source["purchase_price_total"];
	        this.unit_price = source["unit_price"];
	        this.purchased_at = source["purchased_at"];
	    }
	}
	export class ItemDTO {
	    id: string;
	    name: string;
	    unit: string;
	    min_stock_alert: number;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.unit = source["unit"];
	        this.min_stock_alert = source["min_stock_alert"];
	        this.created_at = source["created_at"];
	    }
	}
	export class ProductDTO {
	    id: string;
	    recipe_id: string;
	    quantity_produced: number;
	    unit_cost: number;
	    sale_price: number;
	    produced_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ProductDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.recipe_id = source["recipe_id"];
	        this.quantity_produced = source["quantity_produced"];
	        this.unit_cost = source["unit_cost"];
	        this.sale_price = source["sale_price"];
	        this.produced_at = source["produced_at"];
	    }
	}
	export class RecipeDTO {
	    id: string;
	    name: string;
	    profit_margin_percent: number;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new RecipeDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.profit_margin_percent = source["profit_margin_percent"];
	        this.created_at = source["created_at"];
	    }
	}
	export class RecipeIngredient {
	    recipe_id: string;
	    item_id: string;
	    quantity_needed: number;
	
	    static createFrom(source: any = {}) {
	        return new RecipeIngredient(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.recipe_id = source["recipe_id"];
	        this.item_id = source["item_id"];
	        this.quantity_needed = source["quantity_needed"];
	    }
	}
	export class SaleDTO {
	    id: string;
	    product_id: string;
	    quantity_sold: number;
	    unit_price: number;
	    total_price: number;
	    sold_at: string;
	
	    static createFrom(source: any = {}) {
	        return new SaleDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.product_id = source["product_id"];
	        this.quantity_sold = source["quantity_sold"];
	        this.unit_price = source["unit_price"];
	        this.total_price = source["total_price"];
	        this.sold_at = source["sold_at"];
	    }
	}

}

